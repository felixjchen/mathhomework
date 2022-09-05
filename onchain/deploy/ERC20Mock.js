module.exports = async ({ getNamedAccounts, deployments }) => {
  const { deploy } = deployments;
  const { deployer } = await getNamedAccounts();
  await deploy("ERC20Mock", {
    from: deployer,
    args: ["FRANKTOKEY", "FOK2"],
    log: true,
  });
};
module.exports.tags = ["ERC20Mock"];
