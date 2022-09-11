// We require the Hardhat Runtime Environment explicitly here. This is optional
// but useful for running the script in a standalone fashion through `node <script>`.
//
// You can also run a script with `npx hardhat run <script>`. If you do that, Hardhat
// will compile your contracts, add the Hardhat Runtime Environment's members to the
// global scope, and execute the script.
const hre = require("hardhat");
const { ethers } = require("ethers");

// const WMATIC_ADDRESS = "0x9c3C9283D3e44854697Cd22D3Faa240Cfb032889";
const QS_FACTORY_ADDRESS = "0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32";
const QS_ROUTER_ADDRESS = "0xa5E0829CaCEd8fFDD4De3c43696c57F7D7A678ff";

async function main() {
  const [signer] = await hre.ethers.getSigners();
  // const WMATICFACTORY = await hre.ethers.getContractFactory("WMATIC");
  // const WMATIC = WMATICFACTORY.attach(WMATIC_ADDRESS);

  const ERC20MockFactory = await hre.ethers.getContractFactory("ERC20Mock");
  const FOK = await ERC20MockFactory.deploy("Frank Tokey", "FOK");
  await FOK.deployed();
  const FOK_ADDRESS = FOK.address;
  console.log(`FOK deployed to ${FOK_ADDRESS}`);

  const COK = await ERC20MockFactory.deploy("Cory Tokey", "COK");
  await COK.deployed();
  const COK_ADDRESS = COK.address;
  console.log(`COK deployed to ${COK_ADDRESS}`);

  const UniswapV2FactoryFactory = await hre.ethers.getContractFactory(
    "contracts/uniswap_v2_core/UniswapV2Factory.sol:UniswapV2Factory"
  );
  const QS_FACTORY = UniswapV2FactoryFactory.attach(QS_FACTORY_ADDRESS);

  const QS_PAIR_ADDRESS = (
    await (await QS_FACTORY.createPair(COK_ADDRESS, FOK_ADDRESS)).wait()
  ).events[0].args[2];

  const UniswapV2PairFactory = await hre.ethers.getContractFactory(
    "UniswapV2Pair"
  );

  console.log({ QS_PAIR_ADDRESS });

  const QS_PAIR = UniswapV2PairFactory.attach(QS_PAIR_ADDRESS);

  await FOK.mint(signer.address, ethers.utils.parseEther("99999"));
  await FOK.mint(QS_PAIR_ADDRESS, ethers.utils.parseEther("3"));
  await COK.mint(QS_PAIR_ADDRESS, ethers.utils.parseEther("3"));

  await (await QS_PAIR.sync()).wait();

  await FOK.approve(QS_ROUTER_ADDRESS, ethers.constants.MaxUint256);
  await COK.approve(QS_ROUTER_ADDRESS, ethers.constants.MaxUint256);
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
