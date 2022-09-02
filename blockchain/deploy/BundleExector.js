module.exports = async ({ getNamedAccounts, deployments }) => {
  const { deploy } = deployments;
  const { deployer } = await getNamedAccounts();
  await deploy("FlashBotsMultiCall", {
    from: deployer,
    args: ["0xF479cBA4257371790263Fe9d8F78D9C2c99f1837"],
    log: true,
  });
};
module.exports.tags = ["MyContract"];
