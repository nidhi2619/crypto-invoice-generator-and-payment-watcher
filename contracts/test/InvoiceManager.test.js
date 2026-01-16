const { expect } = require("chai");
const { ethers } = require("hardhat");
const { time } = require("@nomicfoundation/hardhat-network-helpers");

describe("InvoiceManager", function () {
  let invoiceManager;
  let owner, merchant, payer, other;
  
  beforeEach(async function () {
    [owner, merchant, payer, other] = await ethers.getSigners();
    
    const InvoiceManager = await ethers.getContractFactory("InvoiceManager");
    invoiceManager = await InvoiceManager.deploy();
    await invoiceManager.waitForDeployment();
  });
  
  describe("Invoice Creation", function () {
    it("Should create invoice with correct details", async function () {
      const amountWei = ethers.parseEther("0.1");
      const expiresAt = Math.floor(Date.now() / 1000) + 3600; // 1 hour from now
      
      const tx = await invoiceManager.createInvoice(merchant.address, amountWei, expiresAt);
      const receipt = await tx.wait();
      
      // Check event emission
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'InvoiceCreated');
      expect(event).to.not.be.undefined;
      expect(event.args.invoiceId).to.equal(1);
      expect(event.args.merchant).to.equal(merchant.address);
      expect(event.args.amountWei).to.equal(amountWei);
      
      // Check stored invoice
      const invoice = await invoiceManager.getInvoice(1);
      expect(invoice.merchant).to.equal(merchant.address);
      expect(invoice.amountWei).to.equal(amountWei);
      expect(invoice.paid).to.be.false;
    });
    
    it("Should reject invoice creation from non-owner", async function () {
      const amountWei = ethers.parseEther("0.1");
      const expiresAt = Math.floor(Date.now() / 1000) + 3600;
      
      await expect(
        invoiceManager.connect(other).createInvoice(merchant.address, amountWei, expiresAt)
      ).to.be.revertedWithCustomError(invoiceManager, "OwnableUnauthorizedAccount");
    });
    
    it("Should reject zero amount", async function () {
      const expiresAt = Math.floor(Date.now() / 1000) + 3600;
      
      await expect(
        invoiceManager.createInvoice(merchant.address, 0, expiresAt)
      ).to.be.revertedWith("Amount must be greater than 0");
    });
    
    it("Should reject past expiry", async function () {
      const amountWei = ethers.parseEther("0.1");
      const pastTime = Math.floor(Date.now() / 1000) - 3600;
      
      await expect(
        invoiceManager.createInvoice(merchant.address, amountWei, pastTime)
      ).to.be.revertedWith("Expiry must be in future");
    });
  });
  
  describe("Invoice Payment", function () {
    let invoiceId, amountWei, expiresAt;
    
    beforeEach(async function () {
      amountWei = ethers.parseEther("0.1");
      expiresAt = Math.floor(Date.now() / 1000) + 3600;
      
      const tx = await invoiceManager.createInvoice(merchant.address, amountWei, expiresAt);
      const receipt = await tx.wait();
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'InvoiceCreated');
      invoiceId = event.args.invoiceId;
    });
    
    it("Should accept correct payment", async function () {
      const merchantBalanceBefore = await ethers.provider.getBalance(merchant.address);
      
      const tx = await invoiceManager.connect(payer).payInvoice(invoiceId, { value: amountWei });
      const receipt = await tx.wait();
      
      // Check event
      const event = receipt.logs.find(log => log.fragment && log.fragment.name === 'InvoicePaid');
      expect(event).to.not.be.undefined;
      expect(event.args.invoiceId).to.equal(invoiceId);
      expect(event.args.payer).to.equal(payer.address);
      
      // Check invoice updated
      const invoice = await invoiceManager.getInvoice(invoiceId);
      expect(invoice.paid).to.be.true;
      expect(invoice.payer).to.equal(payer.address);
      
      // Check merchant received payment
      const merchantBalanceAfter = await ethers.provider.getBalance(merchant.address);
      expect(merchantBalanceAfter - merchantBalanceBefore).to.equal(amountWei);
    });
    
    it("Should reject incorrect payment amount", async function () {
      const wrongAmount = ethers.parseEther("0.05");
      
      await expect(
        invoiceManager.connect(payer).payInvoice(invoiceId, { value: wrongAmount })
      ).to.be.revertedWith("Incorrect payment amount");
    });
    
    it("Should reject payment for non-existent invoice", async function () {
      await expect(
        invoiceManager.connect(payer).payInvoice(999, { value: amountWei })
      ).to.be.revertedWith("Invoice does not exist");
    });
    
    it("Should reject double payment", async function () {
      await invoiceManager.connect(payer).payInvoice(invoiceId, { value: amountWei });
      
      await expect(
        invoiceManager.connect(other).payInvoice(invoiceId, { value: amountWei })
      ).to.be.revertedWith("Invoice already paid");
    });
    
    it("Should reject payment for expired invoice", async function () {
      // Fast forward time past expiry
      await time.increaseTo(expiresAt + 1);
      
      await expect(
        invoiceManager.connect(payer).payInvoice(invoiceId, { value: amountWei })
      ).to.be.revertedWith("Invoice expired");
    });
  });
});
