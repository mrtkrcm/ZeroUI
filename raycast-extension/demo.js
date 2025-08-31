#!/usr/bin/env node

/**
 * ZeroUI Raycast Extension Demo
 *
 * This script demonstrates how the Raycast extension works
 * by simulating the commands that would be executed.
 */

const { exec } = require('child_process');
const path = require('path');

const ZEROUIPATH = '/Users/m/code/muka-hq/zeroui/build/zeroui';

console.log('🎯 ZeroUI Raycast Extension Demo');
console.log('=================================');
console.log('');

async function runCommand(command, description) {
  return new Promise((resolve) => {
    console.log(`📋 ${description}`);
    console.log(`   Command: ${command}`);
    console.log('');

    exec(`${ZEROUIPATH} ${command}`, (error, stdout, stderr) => {
      if (error) {
        console.log(`❌ Error: ${error.message}`);
      } else if (stderr) {
        console.log(`⚠️  Warning: ${stderr}`);
      } else {
        console.log(`✅ Output:`);
        console.log(stdout);
      }
      console.log('─'.repeat(50));
      console.log('');
      resolve();
    });
  });
}

async function main() {
  try {
    // Check if ZeroUI exists
    const fs = require('fs');
    if (!fs.existsSync(ZEROUIPATH)) {
      console.log('❌ ZeroUI binary not found!');
      console.log(`   Expected at: ${ZEROUIPATH}`);
      console.log('   Please build ZeroUI first:');
      console.log('   cd /Users/m/code/muka-hq/zeroui && make build');
      return;
    }

    console.log('✅ ZeroUI binary found');
    console.log('');

    // Demonstrate commands
    await runCommand('list apps', 'List all available applications');

    await runCommand('list values ghostty | head -10', 'Show first 10 configuration values for ghostty');

    await runCommand('list changed ghostty | head -5', 'Show first 5 changed values for ghostty');

    await runCommand('keymap list ghostty | head -5', 'Show first 5 keymaps for ghostty');

    console.log('🎉 Demo complete!');
    console.log('');
    console.log('💡 In Raycast, these commands would appear as beautiful UI:');
    console.log('   • Interactive lists with search');
    console.log('   • Forms for configuration toggles');
    console.log('   • Copy-to-clipboard actions');
    console.log('   • Visual feedback and error handling');

  } catch (error) {
    console.error('❌ Demo failed:', error.message);
  }
}

if (require.main === module) {
  main();
}

module.exports = { runCommand };
