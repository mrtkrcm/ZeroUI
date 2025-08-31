#!/usr/bin/env node

/**
 * Test script to simulate Raycast extension environment and debug path resolution
 */

const fs = require('fs');
const path = require('path');

// Simulate Raycast extension paths
const testDirname = __dirname;
const testCwd = process.cwd();

console.log('🔍 Path Resolution Test');
console.log('=======================');
console.log(`__dirname: ${testDirname}`);
console.log(`process.cwd(): ${testCwd}`);
console.log('');

// Test the same fallback paths as the extension
const fallbackPaths = [
  path.resolve(testDirname, "../zeroui"), // Relative to built extension
  path.resolve(testDirname, "../../zeroui"), // One level up
  path.resolve(testCwd, "zeroui"), // Current working directory
  path.resolve(testCwd, "../build/zeroui"), // Build directory
  path.resolve(testCwd, "build/zeroui"), // Build directory (alternative)
  "/usr/local/bin/zeroui", // System path
  "/opt/homebrew/bin/zeroui", // Homebrew path
];

console.log('Testing fallback paths:');
let foundPath = null;

for (const testPath of fallbackPaths) {
  if (fs.existsSync(testPath)) {
    const stats = fs.statSync(testPath);
    if (stats.isFile() && (stats.mode & parseInt('111', 8))) {
      console.log(`✅ ${testPath} (executable file)`);
      if (!foundPath) {
        foundPath = testPath;
      }
    } else {
      console.log(`⚠️  ${testPath} (${stats.isDirectory() ? 'directory' : 'not executable'})`);
    }
  } else {
    console.log(`❌ ${testPath} (does not exist)`);
  }
}

console.log('');
if (foundPath) {
  console.log(`🎯 Selected path: ${foundPath}`);

  // Test if it's executable
  try {
    const stats = fs.statSync(foundPath);
    const isExecutable = !!(stats.mode & parseInt('111', 8));
    console.log(`📋 Executable: ${isExecutable ? 'Yes' : 'No'}`);
    console.log(`📋 Size: ${stats.size} bytes`);
  } catch (error) {
    console.error(`❌ Cannot stat file: ${error.message}`);
  }

  // Test the command
  console.log('');
  console.log('🧪 Testing ZeroUI command:');
  const { exec } = require('child_process');
  exec(`${foundPath} list apps`, (error, stdout, stderr) => {
    if (error) {
      console.error(`❌ Command failed: ${error.message}`);
    } else {
      console.log('✅ Command successful:');
      console.log(stdout);
    }
  });

} else {
  console.log('❌ No ZeroUI binary found in any fallback path!');
  console.log('');
  console.log('💡 Solutions:');
  console.log('1. Copy zeroui binary to extension directory:');
  console.log(`   cp ../build/zeroui ${testDirname}/`);
  console.log('2. Or set the path in Raycast preferences');
  console.log('3. Or add zeroui to your PATH');
}

console.log('');
console.log('🔧 Test complete');
