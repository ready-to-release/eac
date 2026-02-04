module.exports = {
  default: {
    // Feature files location - uses repo-level specs directory
    paths: ['../../specs/vscode-commit/**/*.feature'],
    // Step definitions
    require: ['features/steps/**/*.ts'],
    // TypeScript support
    requireModule: ['ts-node/register'],
    // Output formats
    format: [
      'progress-bar',
      ['json', 'out/test/cucumber-report.json']
    ],
    // Fail fast on first failure (optional)
    failFast: false,
    // Parallel execution (optional)
    parallel: 1
  }
};
