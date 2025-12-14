import { exec } from "child_process";
import * as fs from "fs";
import * as os from "os";
import * as path from "path";
import { promisify } from "util";

const execAsync = promisify(exec);

async function findBinary() {
  const fallbackPaths = [
    // Development paths - check parent directories
    path.resolve(__dirname, "../../ZeroUI-arm64"),
    path.resolve(__dirname, "../../ZeroUI"),
    path.resolve(__dirname, "../ZeroUI"),
    path.resolve(__dirname, "../../zeroui"),
    path.resolve(process.cwd(), "ZeroUI"),

    // System paths
    "/usr/local/bin/zeroui",
    path.join(os.homedir(), ".local/bin/zeroui"),
  ];

  console.log("Searching for binary in:");
  for (const p of fallbackPaths) {
    const exists = fs.existsSync(p);
    const isFile = exists && fs.statSync(p).isFile();
    console.log(
      `  ${p}: ${exists ? (isFile ? "FOUND (File)" : "FOUND (Dir)") : "Not found"}`,
    );
    if (isFile) return p;
  }
  return null;
}

function cleanText(text: string): string {
  return (
    text
      .split("\n")
      .filter((line) => !line.match(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/)) // Filter logic from utils.ts
      .join("\n")
      // eslint-disable-next-line no-control-regex
      .replace(/\x1B\[[0-9;]*[a-zA-Z]/g, "") // Strip ANSI
      .trim()
  );
}

async function run() {
  console.log("--- Starting E2E Test ---");

  const binaryPath = await findBinary();
  if (!binaryPath) {
    console.error("CRITICAL: Binary not found!");
    process.exit(1);
  }

  console.log(`\nUsing binary: ${binaryPath}`);

  try {
    console.log("Running 'list apps'...");
    // Emulate utils.ts CWD logic
    const projectRoot = path.resolve(__dirname, "../../");
    console.log(`Setting CWD to: ${projectRoot}`);

    const { stdout, stderr } = await execAsync(`"${binaryPath}" list apps`, {
      cwd: projectRoot,
    });

    console.log("\n--- RAW STDOUT ---");
    console.log(stdout);
    console.log("------------------");

    console.log("\n--- RAW STDERR ---");
    console.log(stderr);
    console.log("------------------");

    const cleanedOutput = cleanText(stdout);
    const cleanedStderr = cleanText(stderr);
    const output = cleanedOutput || cleanedStderr;

    console.log("\n--- CLEANED OUTPUT ---");
    console.log(output);
    console.log("----------------------");

    console.log("\n--- PARSING ---");
    const lines = output.split("\n");
    const apps: string[] = [];

    for (const line of lines) {
      const trimmed = line.trim();
      console.log(`Processing line: '${line}' (trimmed: '${trimmed}')`);

      // Regex from utils.ts
      if (trimmed && trimmed.match(/^[•\-*]/)) {
        const match = trimmed.match(/^[•\-*]\s*([\w.-]+)/);
        if (match) {
          console.log(`  MATCHED: ${match[1]}`);
          apps.push(match[1]);
        } else {
          console.log(`  FAILED MATCH regex on bullet line`);
        }
      } else {
        console.log(`  ignored (no bullet)`);
      }
    }

    console.log(`\nFound ${apps.length} apps: ${apps.join(", ")}`);

    if (apps.length === 0) {
      console.error("FAIL: Parsed 0 apps!");
    } else {
      console.log("SUCCESS: Apps parsed correctly.");
    }
  } catch (err) {
    console.error("Execution failed:", err);
  }
}

run();
