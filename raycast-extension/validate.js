#!/usr/bin/env node

/**
 * ZeroUI Raycast Extension Validator
 *
 * Validates that the Raycast extension can communicate with ZeroUI
 */

const { exec } = require('child_process');
const path = require('path');

// Try multiple possible locations for ZeroUI binary
function findZeroUIPath() {
  const fs = require('fs');
  const path = require('path');

  const possiblePaths = [
    path.join(__dirname, 'zeroui'), // Extension directory
    path.join(__dirname, '..', 'build', 'zeroui'), // Build directory
    path.join(__dirname, '..', 'zeroui'), // Project root
    '/usr/local/bin/zeroui', // System path
    '/opt/homebrew/bin/zeroui', // Homebrew path
  ];

  for (const testPath of possiblePaths) {
    if (fs.existsSync(testPath)) {
      console.log(`✅ Found ZeroUI binary at: ${testPath}`);
      return testPath;
    }
  }

  console.log('❌ ZeroUI binary not found in any of these locations:');
  possiblePaths.forEach(p => console.log(`   - ${p}`));
  return null;
}

const ZEROUIPATH = findZeroUIPath();

console.log('🔍 ZeroUI Raycast Extension Validator');
console.log('=====================================');
console.log('');

async function validateCommand(command, description, expectedPattern = null) {
  return new Promise((resolve) => {
    console.log(`🔍 Testing: ${description}`);

    exec(`${ZEROUIPATH} ${command}`, (error, stdout, stderr) => {
      let success = false;
      let message = '';

      if (error) {
        message = `❌ Failed: ${error.message}`;
      } else if (stderr && stderr.trim()) {
        message = `⚠️  Warning: ${stderr.trim()}`;
      } else if (expectedPattern && !stdout.includes(expectedPattern)) {
        message = `❌ Unexpected output format`;
      } else {
        success = true;
        message = `✅ Success: ${stdout.split('\n')[0]}...`;
      }

      console.log(`   ${message}`);
      console.log('');

      resolve({ success, command, description });
    });
  });
}

async function main() {
  const results = [];

  try {
    // Check if ZeroUI exists
    if (!ZEROUIPATH) {
      console.log('❌ ZeroUI binary not found!');
      console.log('   Please build ZeroUI first:');
      console.log('   cd /Users/m/code/muka-hq/zeroui && make build');
      console.log('   Or copy the binary to the raycast-extension directory');
      return;
    }

    console.log('✅ ZeroUI binary found');
    console.log('');

    // Test basic connectivity
    results.push(await validateCommand('--help', 'CLI help system', 'ZeroUI'));

    // Test core functionality
    results.push(await validateCommand('list apps', 'List applications', 'ghostty'));

    results.push(await validateCommand('list values ghostty', 'List configuration values', ':'));

    results.push(await validateCommand('list changed ghostty', 'List changed values', '(default:'));

    results.push(await validateCommand('keymap list ghostty', 'List keymaps', '→'));

    // Test error handling
    results.push(await validateCommand('list values nonexistent', 'Error handling for invalid app'));

    console.log('📊 Validation Results:');
    console.log('=====================');

    const passed = results.filter(r => r.success).length;
    const total = results.length;

    results.forEach(result => {
      const icon = result.success ? '✅' : '❌';
      console.log(`${icon} ${result.description}`);
    });

    console.log('');
    console.log(`📈 Score: ${passed}/${total} tests passed`);

    if (passed === total) {
      console.log('🎉 All validations passed! Raycast extension is ready.');
      console.log('');
      console.log('🚀 Next steps:');
      console.log('   1. Install the extension in Raycast');
      console.log('   2. Run: raycast-extension/install.sh');
      console.log('   3. Import extension in Raycast');
      console.log('   4. Start using ZeroUI commands!');
    } else {
      console.log('⚠️  Some validations failed. Please check ZeroUI installation.');
    }

  } catch (error) {
    console.error('❌ Validation failed:', error.message);
  }
}

if (require.main === module) {
  main();
}
