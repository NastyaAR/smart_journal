const AcademicMeritToken = artifacts.require("AcademicMeritToken");

module.exports = function (deployer) {
  deployer.deploy(AcademicMeritToken);
};
