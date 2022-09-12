// We require the Hardhat Runtime Environment explicitly here. This is optional
// but useful for running the script in a standalone fashion through `node <script>`.
//
// You can also run a script with `npx hardhat run <script>`. If you do that, Hardhat
// will compile your contracts, add the Hardhat Runtime Environment's members to the
// global scope, and execute the script.
const hre = require("hardhat");
const { ethers } = require("ethers");
const ROUTER_ABI = require("./abi/router02_abi.json");

// const WMATIC_ADDRESS = "0x9c3C9283D3e44854697Cd22D3Faa240Cfb032889";
const FOK_ADDRESS = "0xa0A7B960D5bdEb3bFbe8aaf461359aacd3EcAbE7";
const COK_ADDRESS = "0x72Bec2B4423D3489C5B7C9B1eE766E5411dBFEAd";
const QS_ROUTER_ADDRESS = "0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff";

async function main() {
  // const WMATICFACTORY = await hre.ethers.getContractFactory("WMATIC");
  // const WMATIC = WMATICFACTORY.attach(WMATIC_ADDRESS);

  const [signer] = await hre.ethers.getSigners();

  ROUTER_FACTORY = new ethers.Contract(
    QS_ROUTER_ADDRESS,
    ROUTER_ABI.abi,
    signer
  );

  ROUTER = ROUTER_FACTORY.attach(QS_ROUTER_ADDRESS);

  amountIn = ethers.utils.parseEther("1");
  amountOutMin = ethers.utils.parseEther("0");
  path = [FOK_ADDRESS, COK_ADDRESS];
  to = signer.address;
  deadline = 1694461501;

  txn = await ROUTER.swapExactTokensForTokens(
    amountIn,
    amountOutMin,
    path,
    to,
    deadline
  );

  console.log("TXN HASH:", txn.hash);
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
