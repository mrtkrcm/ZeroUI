#!/usr/bin/env node

/**
 * Capture Extension Logs - Simulates ZeroUI extension startup and logs all console output
 */

// Mock Raycast API
const mockRaycastAPI = {
  getPreferenceValues: () => ({
    zerouiPath: '',
    timeout: '30000',
    enableCache: true,
    cacheDuration: '300000'
  })
};

// Capture console logs
const originalConsoleLog = console.log;
const originalConsoleWarn = console.warn;
const originalConsoleError = console.error;

const capturedLogs = [];

console.log = (...args) => {
  capturedLogs.push(`[LOG] ${new Date().toISOString()}: ${args.join(' ')}`);
  originalConsoleLog(...args);
};

console.warn = (...args) => {
  capturedLogs.push(`[WARN] ${new Date().toISOString()}: ${args.join(' ')}`);
  originalConsoleWarn(...args);
};

console.error = (...args) => {
  capturedLogs.push(`[ERROR] ${new Date().toISOString()}: ${args.join(' ')}`);
  originalConsoleError(...args);
};

// Simulate extension startup
async function simulateExtensionStartup() {
  console.log('🚀 Simulating ZeroUI Raycast Extension Startup...');

  // Mock the environment that would be present in Raycast
  const __dirname = '/Users/m/code/muka-hq/zeroui/raycast-extension';
  const processCwd = '/Users/m/code/muka-hq/zeroui/raycast-extension';

  // Import our actual utils (this will trigger the path resolution logging)
  try {
    console.log('📦 Loading ZeroUI utilities...');

    // Simulate the path resolution that happens in utils.ts constructor
    const fs = require('fs');
    const path = require('path');
    const { getPreferenceValues } = mockRaycastAPI;

    console.log('🔧 Initializing ZeroUI class...');

    // Simulate the path resolution logic from utils.ts
    const preferences = getPreferenceValues();
    console.log('⚙️  Preferences loaded:', JSON.stringify(preferences, null, 2));

    // Try multiple fallback locations (same as in utils.ts)
    const fallbackPaths = [
      path.resolve(__dirname, "../zeroui"), // Relative to built extension
      path.resolve(__dirname, "../../zeroui"), // One level up
      path.resolve(processCwd, "zeroui"), // Current working directory
      path.resolve(processCwd, "../build/zeroui"), // Build directory
      path.resolve(processCwd, "build/zeroui"), // Build directory (alternative)
      "/usr/local/bin/zeroui", // System path
      "/opt/homebrew/bin/zeroui", // Homebrew path
    ];

    console.log('🔍 Testing fallback paths for ZeroUI binary:');

    let zerouiPath = null;
    for (const testPath of fallbackPaths) {
      console.log(`Testing ZeroUI path: ${testPath}`);
      if (fs.existsSync(testPath)) {
        const stats = fs.statSync(testPath);
        if (stats.isFile() && (stats.mode & parseInt('111', 8))) {
          console.log(`✅ Found executable ZeroUI binary at: ${testPath}`);
          zerouiPath = testPath;
          break;
        } else {
          console.log(`⚠️  Path exists but is not an executable file: ${testPath} (${stats.isDirectory() ? 'directory' : 'not executable'})`);
        }
      } else {
        console.log(`❌ Path does not exist: ${testPath}`);
      }
    }

    if (!zerouiPath) {
      zerouiPath = path.resolve(__dirname, "../zeroui");
      console.warn(`No ZeroUI binary found in fallback paths, using default: ${zerouiPath}`);
    }

    console.log(`ZeroUI binary path resolved to: ${zerouiPath}`);
    console.log(`__dirname: ${__dirname}`);
    console.log(`process.cwd(): ${processCwd}`);
    console.log(`Current directory contents:`, fs.readdirSync(path.dirname(zerouiPath)).slice(0, 10));

    // Double-check if the binary actually exists at the resolved path
    if (!fs.existsSync(zerouiPath)) {
      console.error(`CRITICAL: ZeroUI binary not found at resolved path: ${zerouiPath}`);
      console.error(`This will cause 'no applications found' error!`);
    } else {
      console.log(`✅ ZeroUI binary exists at resolved path`);
    }

    // Test the binary
    console.log('🧪 Testing ZeroUI binary functionality...');

    const { execSync } = require('child_process');

    try {
      const versionOutput = execSync(`${zerouiPath} version`, { encoding: 'utf8', timeout: 5000 });
      console.log('✅ ZeroUI version check successful:', versionOutput.trim());
    } catch (error) {
      console.error('❌ ZeroUI version check failed:', error.message);
    }

    try {
      const appsOutput = execSync(`${zerouiPath} list apps`, { encoding: 'utf8', timeout: 5000 });
      console.log('✅ ZeroUI list apps successful:');
      console.log(appsOutput);
    } catch (error) {
      console.error('❌ ZeroUI list apps failed:', error.message);
    }

    console.log('📊 Cache statistics simulation:');
    console.log('✅ Cache hits: 0');
    console.log('✅ Cache misses: 0');
    console.log('✅ Hit rate: 0%');
    console.log('✅ Total requests: 0');

  } catch (error) {
    console.error('❌ Extension startup simulation failed:', error.message);
    console.error('Stack trace:', error.stack);
  }
}

// Run the simulation
simulateExtensionStartup().then(() => {
  console.log('\n📋 LOG CAPTURE COMPLETE');
  console.log('='.repeat(50));

  const finalOutput = capturedLogs.join('\n');

  // Write to file
  require('fs').writeFileSync('extension-logs.txt', finalOutput);

  // Copy to clipboard (macOS)
  try {
    require('child_process').execSync(`echo "${finalOutput.replace(/"/g, '\\"')}" | pbcopy`);
    console.log('✅ Logs copied to clipboard!');
  } catch (clipboardError) {
    console.log('⚠️ Could not copy to clipboard, but logs saved to file');
  }

  console.log('📄 Full logs also saved to: extension-logs.txt');
  console.log('\n🔍 CAPTURED LOGS:');
  console.log(finalOutput);
}).catch(error => {
  console.error('❌ Log capture failed:', error);
});
