const hre = require("hardhat");

async function main() {
  console.log("Deploying InvoiceManager contract to Sepolia...");

  const [deployer] = await hre.ethers.getSigners();
  console.log("Deploying with account:", deployer.address);
  
  const balance = await hre.ethers.provider.getBalance(deployer.address);
  console.log("Account balance:", hre.ethers.formatEther(balance), "ETH");

  // Deploy contract
  const InvoiceManager = await hre.ethers.getContractFactory("InvoiceManager");
  const invoiceManager = await InvoiceManager.deploy();
  
  await invoiceManager.waitForDeployment();
  
  const contractAddress = await invoiceManager.getAddress();
  console.log("✅ InvoiceManager deployed to:", contractAddress);
  
  console.log("\n📝 Update your backend/.env with:");
  console.log(`CONTRACT_ADDRESS=${contractAddress}`);
  
  console.log("\n⏳ Waiting 30 seconds before verification...");
  await new Promise(resolve => setTimeout(resolve, 30000));
  
  // Verify on Etherscan
  try {
    console.log("Verifying contract on Etherscan...");
    await hre.run("verify:verify", {
      address: contractAddress,
      constructorArguments: [],
    });
    console.log("✅ Contract verified on Etherscan");
  } catch (error) {
    console.log("⚠️  Verification failed:", error.message);
  }
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
