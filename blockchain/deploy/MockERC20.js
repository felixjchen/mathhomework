module.exports = async ({ getNamedAccounts, deployments }) => {
  const { deploy } = deployments;
  const { deployer } = await getNamedAccounts();
  await deploy("ERC20Mock", {
    from: deployer,
    args: ["Frank Token", "FOK"],
    log: true,
  });
};
module.exports.tags = ["MyContract"];
