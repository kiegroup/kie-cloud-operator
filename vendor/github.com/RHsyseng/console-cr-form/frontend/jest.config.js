// eslint-disable-next-line no-undef
module.exports = {
  automock: false,
  setupFiles: ["./jest.setup"],
  moduleNameMapper: {
    "\\.css$": require.resolve("./__mocks__/styleMock")
  }
};
