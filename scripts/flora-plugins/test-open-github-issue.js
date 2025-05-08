// Test script for the Open GitHub Issue Flora plugin
// This script demonstrates how to use the plugin programmatically

// Import the plugin
const openGitHubIssuePlugin = require('../../.plugins/open-github-issue');

// Display plugin metadata
console.log('Plugin Metadata:');
console.log(`Name: ${openGitHubIssuePlugin.metadata.name}`);
console.log(`Description: ${openGitHubIssuePlugin.metadata.description}`);
console.log(`Version: ${openGitHubIssuePlugin.metadata.version}`);
console.log(`Author: ${openGitHubIssuePlugin.metadata.author}`);
console.log('Commands:');
openGitHubIssuePlugin.metadata.commands.forEach(command => {
  console.log(`- ${command.name}: ${command.description}`);
  console.log(`  Usage: ${command.usage}`);
});
console.log('');

// Test the plugin with a valid issue number
async function testValidIssue() {
  console.log('Testing with valid issue number:');
  const issueNumber = 123;
  console.log(`Opening GitHub issue #${issueNumber}...`);

  try {
    const result = await openGitHubIssuePlugin.commands.openGitHubIssue(issueNumber);
    console.log('Result:', result);
  } catch (error) {
    console.error('Error:', error.message);
  }
}

// Test the plugin with a valid issue number and custom IDE
async function testValidIssueWithCustomIde() {
  console.log('\nTesting with valid issue number and custom IDE:');
  const issueNumber = 456;
  const options = { ide: 'idea' };
  console.log(`Opening GitHub issue #${issueNumber} with IDE: ${options.ide}...`);

  try {
    const result = await openGitHubIssuePlugin.commands.openGitHubIssue(issueNumber, options);
    console.log('Result:', result);
  } catch (error) {
    console.error('Error:', error.message);
  }
}

// Test the plugin with an invalid issue number
async function testInvalidIssue() {
  console.log('\nTesting with invalid issue number:');
  const issueNumber = 'abc';
  console.log(`Opening GitHub issue #${issueNumber}...`);

  try {
    const result = await openGitHubIssuePlugin.commands.openGitHubIssue(issueNumber);
    console.log('Result:', result);
  } catch (error) {
    console.error('Error:', error.message);
  }
}

// Run the tests
async function runTests() {
  await testValidIssue();
  await testValidIssueWithCustomIde();
  await testInvalidIssue();

  console.log('\nTests completed.');
}

// Execute the tests
runTests().catch(error => {
  console.error('Test execution error:', error.message);
});

// Example of how to use the plugin in Flora
console.log('\nExample usage in Flora:');
console.log('// Basic usage');
console.log('Flora.commands.openGitHubIssue(123);');
console.log('\n// With custom IDE');
console.log('Flora.commands.openGitHubIssue(123, { ide: "idea" });');

// Example of how to use the plugin with Junie AI
console.log('\nExample usage with Junie AI:');
console.log('// Basic usage');
console.log('User: "Junie, can you open GitHub issue 123?"');
console.log('Junie: "I\'ll open GitHub issue 123 for you."');
console.log('(Junie uses Flora.commands.openGitHubIssue(123) to open the issue)');
console.log('\n// With custom IDE');
console.log('User: "Junie, can you open GitHub issue 123 in IntelliJ IDEA?"');
console.log('Junie: "I\'ll open GitHub issue 123 in IntelliJ IDEA for you."');
console.log('(Junie uses Flora.commands.openGitHubIssue(123, { ide: "idea" }) to open the issue)');
