#!/usr/bin/env node

/**
 * Comprehensive ZeroUI Raycast Extension Diagnostic
 * Captures all relevant information for debugging extension issues
 */

const fs = require('fs');
const path = require('path');
const { exec, execSync } = require('child_process');

const output = [];

// Header
output.push('='.repeat(60));
output.push('🔍 ZeroUI Raycast Extension Comprehensive Diagnostic');
output.push('='.repeat(60));
output.push(`Generated: ${new Date().toISOString()}`);
output.push('');

async function runCommand(cmd, description) {
  try {
    const result = execSync(cmd, { encoding: 'utf8', timeout: 10000 });
    output.push(`✅ ${description}:`);
    output.push(result.trim());
    output.push('');
  } catch (error) {
    output.push(`❌ ${description} (FAILED):`);
    output.push(`Error: ${error.message}`);
    output.push('');
  }
}

async function checkFile(filePath, description) {
  try {
    const stats = fs.statSync(filePath);
    const isExecutable = !!(stats.mode & parseInt('111', 8));
    output.push(`✅ ${description}:`);
    output.push(`  Path: ${filePath}`);
    output.push(`  Size: ${stats.size} bytes`);
    output.push(`  Executable: ${isExecutable}`);
    output.push(`  Modified: ${stats.mtime.toISOString()}`);
    output.push('');
  } catch (error) {
    output.push(`❌ ${description} (NOT FOUND):`);
    output.push(`  Path: ${filePath}`);
    output.push(`  Error: ${error.message}`);
    output.push('');
  }
}

async function main() {
  // Basic system info
  output.push('📊 SYSTEM INFORMATION:');
  await runCommand('uname -a', 'System info');
  await runCommand('node --version', 'Node.js version');
  await runCommand('npm --version', 'NPM version');

  // Directory structure
  output.push('📁 DIRECTORY STRUCTURE:');
  output.push(`Current directory: ${process.cwd()}`);
  output.push(`__dirname: ${__dirname}`);
  output.push(`process.cwd(): ${process.cwd()}`);
  output.push('');

  // Check extension files
  output.push('📄 EXTENSION FILES:');
  await checkFile('package.json', 'Package manifest');
  await checkFile('tsconfig.json', 'TypeScript config');
  await checkFile('src/utils.ts', 'Main utilities');
  await checkFile('src/list-apps.tsx', 'List apps component');

  // Check ZeroUI binary
  output.push('🔧 ZeroUI BINARY:');
  await checkFile('zeroui', 'ZeroUI executable');

  // Test ZeroUI functionality
  output.push('🧪 ZeroUI FUNCTIONALITY TESTS:');
  await runCommand('./zeroui --version 2>/dev/null || ./zeroui --help | head -3', 'ZeroUI version/help');
  await runCommand('./zeroui list apps', 'List applications');
  await runCommand('./zeroui list values ghostty 2>/dev/null | head -5 || echo "Failed to list ghostty values"', 'Test ghostty config');

  // Check build status
  output.push('🔨 BUILD STATUS:');
  const builtFiles = fs.readdirSync('.').filter(f => f.endsWith('.js') && !f.includes('node_modules'));
  output.push(`Built JS files: ${builtFiles.join(', ') || 'None found'}`);
  output.push('');

  // Node modules status
  output.push('📦 NODE MODULES:');
  try {
    const nodeModules = fs.readdirSync('node_modules');
    output.push(`Modules count: ${nodeModules.length}`);
    const keyModules = ['@raycast/api', '@raycast/utils', 'react'];
    keyModules.forEach(mod => {
      if (nodeModules.includes(mod)) {
        output.push(`✅ ${mod}: Installed`);
      } else {
        output.push(`❌ ${mod}: Missing`);
      }
    });
  } catch (error) {
    output.push(`❌ Node modules error: ${error.message}`);
  }
  output.push('');

  // Raycast build check
  output.push('🎯 RAYCAST BUILD CHECK:');
  await runCommand('npx ray --version 2>/dev/null || echo "Raycast CLI not found"', 'Raycast CLI version');
  await runCommand('find . -name "*.raycast" -o -name "dist" -type d 2>/dev/null || echo "No Raycast build artifacts found"', 'Raycast build artifacts');

  // Recent file changes
  output.push('📝 RECENT CHANGES:');
  await runCommand('find src/ -name "*.ts" -o -name "*.tsx" | head -5 | xargs ls -la', 'Source files timestamps');

  // Generate final output
  const finalOutput = output.join('\n');

  // Write to file
  fs.writeFileSync('diagnostic-report.txt', finalOutput);
  console.log('📋 Diagnostic report saved to: diagnostic-report.txt');
  console.log('');

  // Copy to clipboard
  try {
    execSync(`echo "${finalOutput.replace(/"/g, '\\"')}" | pbcopy`);
    console.log('✅ Report copied to clipboard!');
  } catch (error) {
    console.log('⚠️ Could not copy to clipboard, but report saved to file');
  }

  console.log('');
  console.log('📊 DIAGNOSTIC SUMMARY:');
  console.log(finalOutput);
}

// Run diagnostic
main().catch(console.error);
