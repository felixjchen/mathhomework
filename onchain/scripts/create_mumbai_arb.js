// We require the Hardhat Runtime Environment explicitly here. This is optional
// but useful for running the script in a standalone fashion through `node <script>`.
//
// You can also run a script with `npx hardhat run <script>`. If you do that, Hardhat
// will compile your contracts, add the Hardhat Runtime Environment's members to the
// global scope, and execute the script.
const hre = require("hardhat");
const { ethers } = require("ethers");

const WMATIC_ADDRESS = "0x9c3C9283D3e44854697Cd22D3Faa240Cfb032889";
const QS_FACTORY_ADDRESS = "0x5757371414417b8C6CAad45bAeF941aBc7d3Ab32";
const SS_FACTORY_ADDRESS = "0xc35DADB65012eC5796536bD9864eD8773aBc74C4";

async function main() {
  const WMATICFACTORY = await hre.ethers.getContractFactory("WMATIC");
  const WMATIC = WMATICFACTORY.attach(WMATIC_ADDRESS);

  const ERC20MockFactory = await hre.ethers.getContractFactory("ERC20Mock");
  const ERC20 = await ERC20MockFactory.deploy("Frank Tokey", "FOK");
  await ERC20.deployed();
  const FOK = ERC20.address;

  console.log(`FOK deployed to ${FOK}`);

  const UniswapV2FactoryFactory = await hre.ethers.getContractFactory(
    "contracts/uniswap_v2_core/UniswapV2Factory.sol:UniswapV2Factory"
  );
  const QS_FACTORY = UniswapV2FactoryFactory.attach(QS_FACTORY_ADDRESS);
  const SS_FACTORY = UniswapV2FactoryFactory.attach(SS_FACTORY_ADDRESS);

  const QS_PAIR_ADDRESS = (
    await (await QS_FACTORY.createPair(WMATIC_ADDRESS, FOK)).wait()
  ).events[0].args[2];
  const SS_PAIR_ADDRESS = (
    await (await SS_FACTORY.createPair(WMATIC_ADDRESS, FOK)).wait()
  ).events[0].args[2];

  const UniswapV2PairFactory = await hre.ethers.getContractFactory(
    "UniswapV2Pair"
  );

  console.log({ QS_PAIR_ADDRESS, SS_PAIR_ADDRESS });

  const QS_PAIR = UniswapV2PairFactory.attach(QS_PAIR_ADDRESS);
  const SS_PAIR = UniswapV2PairFactory.attach(SS_PAIR_ADDRESS);

  const [signer] = await hre.ethers.getSigners();
  await ERC20.mint(signer.address, ethers.utils.parseEther("99999"));
  await ERC20.mint(QS_PAIR_ADDRESS, ethers.utils.parseEther("200"));
  await ERC20.mint(SS_PAIR_ADDRESS, ethers.utils.parseEther("0.001"));

  await WMATIC.transfer(QS_PAIR_ADDRESS, ethers.utils.parseEther("0.001"));
  await WMATIC.transfer(SS_PAIR_ADDRESS, ethers.utils.parseEther("0.001"));

  await (await QS_PAIR.sync()).wait();
  await (await SS_PAIR.sync()).wait();
}

// We recommend this pattern to be able to use async/await everywhere
// and properly handle errors.
main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
