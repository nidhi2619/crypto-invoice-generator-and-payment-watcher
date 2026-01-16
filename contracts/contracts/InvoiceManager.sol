// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/access/Ownable.sol";

/**
 * @title InvoiceManager
 * @dev Contract for creating and paying ETH invoices on-chain
 * @notice Invoice creation karne ke liye owner permission chahiye
 * Payment koi bhi kar sakta hai using payInvoice function
 */
contract InvoiceManager is ReentrancyGuard, Ownable {
    
    struct Invoice {
        address merchant;      // Merchant ka address jisko payment milegi
        uint256 amountWei;     // Required payment amount in wei
        uint256 expiresAt;     // Unix timestamp - iske baad payment accept nahi hogi
        bool paid;             // Payment status
        address payer;         // Jisne payment kiya (zero address if unpaid)
    }
    
    // Invoice ID counter - har naye invoice ke liye increment hoga
    uint256 private _nextInvoiceId = 1;
    
    // Mapping: invoiceId => Invoice struct
    mapping(uint256 => Invoice) public invoices;
    
    // Events - Backend watcher in events ko listen karega
    event InvoiceCreated(
        uint256 indexed invoiceId,
        address indexed merchant,
        uint256 amountWei,
        uint256 expiresAt
    );
    
    event InvoicePaid(
        uint256 indexed invoiceId,
        address indexed payer,
        uint256 amountWei
    );
    
    constructor() Ownable(msg.sender) {}
    
    /**
     * @dev Create new invoice on-chain
     * @param merchant Address jisko payment forward hogi
     * @param amountWei Payment amount in wei (exact match required)
     * @param expiresAt Unix timestamp for expiry
     * @return invoiceId Generated invoice ID
     */
    function createInvoice(
        address merchant,
        uint256 amountWei,
        uint256 expiresAt
    ) external onlyOwner returns (uint256) {
        require(merchant != address(0), "Invalid merchant address");
        require(amountWei > 0, "Amount must be greater than 0");
        require(expiresAt > block.timestamp, "Expiry must be in future");
        
        uint256 invoiceId = _nextInvoiceId++;
        
        invoices[invoiceId] = Invoice({
            merchant: merchant,
            amountWei: amountWei,
            expiresAt: expiresAt,
            paid: false,
            payer: address(0)
        });
        
        emit InvoiceCreated(invoiceId, merchant, amountWei, expiresAt);
        
        return invoiceId;
    }
    
    /**
     * @dev Pay an existing invoice
     * @param invoiceId ID of invoice to pay
     * @notice 
     */
    function payInvoice(uint256 invoiceId) external payable nonReentrant {
        Invoice storage invoice = invoices[invoiceId];
        
        // Validations
        require(invoice.merchant != address(0), "Invoice does not exist");
        require(!invoice.paid, "Invoice already paid");
        require(block.timestamp <= invoice.expiresAt, "Invoice expired");
        require(msg.value == invoice.amountWei, "Incorrect payment amount");
        
        // Effects (state changes pehle, external calls baad mein)
        invoice.paid = true;
        invoice.payer = msg.sender;
        
        // Emit event before external call
        emit InvoicePaid(invoiceId, msg.sender, msg.value);
        
        // Interactions - merchant ko ETH forward karo
        (bool success, ) = invoice.merchant.call{value: msg.value}("");
        require(success, "Payment forward failed");
    }
    
    /**
     * @dev Get invoice details
     * @param invoiceId Invoice ID to query
     */
    function getInvoice(uint256 invoiceId) external view returns (
        address merchant,
        uint256 amountWei,
        uint256 expiresAt,
        bool paid,
        address payer
    ) {
        Invoice memory invoice = invoices[invoiceId];
        return (
            invoice.merchant,
            invoice.amountWei,
            invoice.expiresAt,
            invoice.paid,
            invoice.payer
        );
    }
    
    /**
     * @dev Get next invoice ID (for testing/debugging)
     */
    function getNextInvoiceId() external view returns (uint256) {
        return _nextInvoiceId;
    }
}
