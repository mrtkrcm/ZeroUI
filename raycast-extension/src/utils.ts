import { getPreferenceValues } from "@raycast/api";
import { exec } from "child_process";
import * as fs from "fs";
import * as path from "path";
import { promisify } from "util";

const execAsync = promisify(exec);

export interface ZeroUIResult {
  success: boolean;
  output: string;
  error?: string;
  duration?: number;
}

export interface Preferences {
  zerouiPath: string;
  timeout: string;
  enableCache: boolean;
  cacheDuration: string;
}

export interface CacheStats {
  size: number;
  hits: number;
  misses: number;
  hitRate: number;
  totalRequests: number;
}

export interface ErrorContext {
  operation: string;
  app?: string;
  key?: string;
  value?: string;
  timestamp: Date;
  retryCount?: number;
}

export enum LogLevel {
  DEBUG = 0,
  INFO = 1,
  WARN = 2,
  ERROR = 3,
}

export interface LogContext {
  [key: string]: string | number | boolean | null | undefined;
}

export interface LogEntry {
  level: LogLevel;
  message: string;
  context?: LogContext;
  timestamp: Date;
}

interface CacheEntry {
  data: ZeroUIResult;
  timestamp: number;
  ttl: number;
}

export class ZeroUI {
  private zerouiPath: string;
  private timeout: number;
  private cache: Map<string, CacheEntry>;
  private enableCache: boolean;
  private cacheDuration: number;
  private cacheHits: number = 0;
  private cacheMisses: number = 0;
  private errorHistory: ErrorContext[] = [];
  private maxRetries: number = 2;
  private retryDelay: number = 1000; // 1 second
  private maxCacheSize: number = 100; // Maximum cache entries
  private cleanupInterval: number = 300000; // 5 minutes cleanup interval
  private logLevel: LogLevel = LogLevel.INFO;
  private logs: LogEntry[] = [];
  private maxLogs: number = 200;

  // Input validation patterns
  private readonly APP_NAME_PATTERN = /^[a-zA-Z0-9_-]+$/;
  private readonly CONFIG_KEY_PATTERN = /^[a-zA-Z0-9_.-]+$/;
  private readonly MAX_APP_NAME_LENGTH = 50;
  private readonly MAX_KEY_LENGTH = 100;
  private readonly MAX_VALUE_LENGTH = 1000;

  constructor(zerouiPath?: string) {
    // Get preferences from Raycast
    const preferences = getPreferenceValues<Preferences>();

    // Try multiple fallback paths for the ZeroUI binary
    if (zerouiPath) {
      this.zerouiPath = zerouiPath;
    } else if (preferences.zerouiPath) {
      this.zerouiPath = preferences.zerouiPath;
    } else {
      // Try multiple fallback locations
      const fallbackPaths = [
        path.resolve(__dirname, "../zeroui"), // Relative to built extension
        path.resolve(__dirname, "../../zeroui"), // One level up
        path.resolve(process.cwd(), "zeroui"), // Current working directory
        path.resolve(process.cwd(), "../build/zeroui"), // Build directory
        path.resolve(process.cwd(), "build/zeroui"), // Build directory (alternative)
        "/usr/local/bin/zeroui", // System path
        "/opt/homebrew/bin/zeroui", // Homebrew path
      ];

      // Find the first existing binary (must be a file, not directory)
      for (const testPath of fallbackPaths) {
        if (fs.existsSync(testPath)) {
          const stats = fs.statSync(testPath);
          if (stats.isFile() && stats.mode & parseInt("111", 8)) {
            this.zerouiPath = testPath;
            break;
          }
        }
      }

      // If no binary found, use the default relative path (might work in development)
      if (!this.zerouiPath) {
        this.zerouiPath = path.resolve(__dirname, "../zeroui");
        console.warn(`No ZeroUI binary found in fallback paths, using default: ${this.zerouiPath}`);
      }
    }

    // Verify the binary exists and is executable
    if (!fs.existsSync(this.zerouiPath)) {
      console.warn(
        `ZeroUI binary not found at: ${this.zerouiPath}, extension may not work correctly`,
      );
    }
    this.timeout = parseInt(preferences.timeout) || 30000; // 30 seconds default
    this.enableCache = preferences.enableCache ?? true;
    this.cacheDuration = parseInt(preferences.cacheDuration) || 300000; // 5 minutes default
    this.cache = new Map();

    // Start periodic cleanup
    this.startPeriodicCleanup();
  }

  private getCacheKey(command: string, args: string[]): string {
    return `${command}:${args.join(":")}`;
  }

  private getCachedResult(key: string): ZeroUIResult | null {
    if (!this.enableCache) return null;

    const entry = this.cache.get(key);
    if (!entry) {
      this.cacheMisses++;
      return null;
    }

    const now = Date.now();
    if (now - entry.timestamp > entry.ttl) {
      this.cache.delete(key);
      this.cacheMisses++;
      return null;
    }

    this.cacheHits++;
    return entry.data;
  }

  private setCachedResult(key: string, data: ZeroUIResult): void {
    if (!this.enableCache) return;

    // Check cache size limit before adding
    if (this.cache.size >= this.maxCacheSize) {
      this.evictOldEntries();
    }

    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl: this.cacheDuration,
    });
  }

  private evictOldEntries(): void {
    const entries = Array.from(this.cache.entries());

    // Sort by timestamp (oldest first)
    entries.sort((a, b) => a[1].timestamp - b[1].timestamp);

    // Remove oldest 20% of entries
    const entriesToRemove = Math.ceil(entries.length * 0.2);
    for (let i = 0; i < entriesToRemove; i++) {
      this.cache.delete(entries[i][0]);
    }
  }

  private startPeriodicCleanup(): void {
    // Clean up expired entries every 5 minutes
    setInterval(() => {
      this.cleanupExpiredEntries();
    }, this.cleanupInterval);
  }

  private cleanupExpiredEntries(): void {
    const now = Date.now();
    const keysToDelete: string[] = [];

    for (const [key, entry] of this.cache.entries()) {
      if (now - entry.timestamp > entry.ttl) {
        keysToDelete.push(key);
      }
    }

    keysToDelete.forEach((key) => this.cache.delete(key));

    if (keysToDelete.length > 0) {
      console.log(`Cleaned up ${keysToDelete.length} expired cache entries`);
    }
  }

  private recordError(context: ErrorContext): void {
    this.errorHistory.push(context);
    // Keep only last 50 errors
    if (this.errorHistory.length > 50) {
      this.errorHistory = this.errorHistory.slice(-50);
    }
  }

  private async executeCommandWithRetry(
    command: string,
    args: string[],
    retryCount: number = 0,
    useCache: boolean = true,
  ): Promise<ZeroUIResult> {
    try {
      return await this.executeCommandOnce(command, args, useCache);
    } catch (error) {
      const shouldRetry = retryCount < this.maxRetries && this.isRetryableError(error);

      if (shouldRetry) {
        console.warn(`Command failed, retrying (${retryCount + 1}/${this.maxRetries}):`, error);
        await new Promise((resolve) => setTimeout(resolve, this.retryDelay * (retryCount + 1)));
        return this.executeCommandWithRetry(command, args, retryCount + 1, useCache);
      }

      throw error;
    }
  }

  private isRetryableError(error: unknown): boolean {
    if (!(error instanceof Error)) return false;

    const retryablePatterns = [
      "timeout",
      "ECONNRESET",
      "ENOTFOUND",
      "ECONNREFUSED",
      "Temporary failure",
    ];

    return retryablePatterns.some((pattern) =>
      error.message.toLowerCase().includes(pattern.toLowerCase()),
    );
  }

  private async executeCommandOnce(
    command: string,
    args: string[],
    useCache: boolean = true,
  ): Promise<ZeroUIResult> {
    const startTime = Date.now();
    const cacheKey = this.getCacheKey(command, args);

    // Check cache first
    if (useCache) {
      const cachedResult = this.getCachedResult(cacheKey);
      if (cachedResult) {
        return {
          ...cachedResult,
          duration: Date.now() - startTime,
        };
      }
    }

    // Check if binary exists
    if (!fs.existsSync(this.zerouiPath)) {
      const error = new Error(
        `ZeroUI binary not found at: ${this.zerouiPath}. Please check the binary path in Raycast preferences or ensure the binary is properly installed.`,
      );

      console.error(`ZeroUI binary not found at: ${this.zerouiPath}`);
      console.error(`__dirname: ${__dirname}`);
      console.error(`process.cwd(): ${process.cwd()}`);
      console.error(`Available files in extension directory:`);
      try {
        const files = fs.readdirSync(path.dirname(this.zerouiPath));
        console.error("Available files:", files);
      } catch (e) {
        console.error(`Could not list directory: ${e}`);
      }

      this.recordError({
        operation: command,
        timestamp: new Date(),
        retryCount: 0,
      });

      throw error;
    }

    try {
      const fullCommand = `${this.zerouiPath} ${command} ${args.join(" ")}`.trim();
      console.log(`Executing: ${fullCommand}`);

      const { stdout, stderr } = await execAsync(fullCommand, {
        timeout: this.timeout,
        maxBuffer: 1024 * 1024 * 10, // 10MB buffer
        cwd: path.resolve(__dirname, "../../"),
      });

      const duration = Date.now() - startTime;
      const hasError = stderr && stderr.trim().length > 0;
      const output = stdout || stderr || "";

      const result: ZeroUIResult = {
        success: !hasError,
        output: output.trim(),
        error: hasError ? stderr.trim() : undefined,
        duration,
      };

      // Cache successful results
      if (result.success && useCache) {
        this.setCachedResult(cacheKey, result);
      }

      return result;
    } catch (error: unknown) {
      console.error("ZeroUI command failed:", error);
      const duration = Date.now() - startTime;
      const errorMessage = error instanceof Error ? error.message : "Unknown error occurred";

      this.recordError({
        operation: command,
        timestamp: new Date(),
        retryCount: 0,
      });

      // Log the error with duration for debugging
      this.error(`Command failed: ${command}`, {
        duration: duration.toString(),
        error: errorMessage,
      });

      throw new Error(errorMessage);
    }
  }

  async executeCommand(
    command: string,
    args: string[] = [],
    useCache: boolean = true,
  ): Promise<ZeroUIResult> {
    const startTime = Date.now();
    this.debug(`Executing command: ${command}`, { args, useCache });

    try {
      const result = await this.executeCommandWithRetry(command, args, 0, useCache);
      const totalDuration = Date.now() - startTime;

      if (result.success) {
        this.debug(`Command completed successfully`, {
          command,
          args,
          duration: result.duration,
          totalDuration,
        });
      } else {
        this.warn(`Command completed with error`, {
          command,
          args,
          error: result.error,
          duration: result.duration,
          totalDuration,
        });
      }

      return result;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : "Unknown error occurred";
      const totalDuration = Date.now() - startTime;

      this.error(`Command failed completely`, {
        command,
        args,
        error: errorMessage,
        totalDuration,
      });

      return {
        success: false,
        output: "",
        error: errorMessage,
        duration: totalDuration,
      };
    }
  }

  clearCache(): void {
    this.cache.clear();
    this.cacheHits = 0;
    this.cacheMisses = 0;
  }

  getCacheStats(): CacheStats {
    const totalRequests = this.cacheHits + this.cacheMisses;
    const hitRate = totalRequests > 0 ? (this.cacheHits / totalRequests) * 100 : 0;

    return {
      size: this.cache.size,
      hits: this.cacheHits,
      misses: this.cacheMisses,
      hitRate: Math.round(hitRate * 100) / 100, // Round to 2 decimal places
      totalRequests,
    };
  }

  getErrorHistory(): ErrorContext[] {
    return [...this.errorHistory];
  }

  clearErrorHistory(): void {
    this.errorHistory = [];
  }

  // Logging methods
  private log(level: LogLevel, message: string, context?: LogContext): void {
    if (level < this.logLevel) return;

    const logEntry: LogEntry = {
      level,
      message,
      context,
      timestamp: new Date(),
    };

    this.logs.push(logEntry);

    // Keep only the most recent logs
    if (this.logs.length > this.maxLogs) {
      this.logs = this.logs.slice(-this.maxLogs);
    }

    // Also log to console for development
    const levelName = LogLevel[level];
    const contextStr = context ? ` ${JSON.stringify(context)}` : "";
    console.log(`[ZeroUI ${levelName}] ${message}${contextStr}`);
  }

  debug(message: string, context?: LogContext): void {
    this.log(LogLevel.DEBUG, message, context);
  }

  info(message: string, context?: LogContext): void {
    this.log(LogLevel.INFO, message, context);
  }

  warn(message: string, context?: LogContext): void {
    this.log(LogLevel.WARN, message, context);
  }

  error(message: string, context?: LogContext): void {
    this.log(LogLevel.ERROR, message, context);
  }

  getLogs(level?: LogLevel): LogEntry[] {
    if (level === undefined) {
      return [...this.logs];
    }
    return this.logs.filter((log) => log.level >= level);
  }

  setLogLevel(level: LogLevel): void {
    this.logLevel = level;
  }

  clearLogs(): void {
    this.logs = [];
  }

  // Input validation methods
  private validateAppName(app: string): void {
    if (!app || typeof app !== "string") {
      throw new Error("App name is required and must be a string");
    }

    if (app.length === 0) {
      throw new Error("App name cannot be empty");
    }

    if (app.length > this.MAX_APP_NAME_LENGTH) {
      throw new Error(`App name is too long (max ${this.MAX_APP_NAME_LENGTH} characters)`);
    }

    if (!this.APP_NAME_PATTERN.test(app)) {
      throw new Error(
        "App name contains invalid characters. Only letters, numbers, hyphens, and underscores are allowed",
      );
    }
  }

  private validateConfigKey(key: string): void {
    if (!key || typeof key !== "string") {
      throw new Error("Configuration key is required and must be a string");
    }

    if (key.length === 0) {
      throw new Error("Configuration key cannot be empty");
    }

    if (key.length > this.MAX_KEY_LENGTH) {
      throw new Error(`Configuration key is too long (max ${this.MAX_KEY_LENGTH} characters)`);
    }

    if (!this.CONFIG_KEY_PATTERN.test(key)) {
      throw new Error(
        "Configuration key contains invalid characters. Only letters, numbers, dots, hyphens, and underscores are allowed",
      );
    }
  }

  private sanitizeValue(value: string): string {
    if (!value || typeof value !== "string") {
      return "";
    }

    // Trim whitespace
    let sanitized = value.trim();

    // Limit length
    if (sanitized.length > this.MAX_VALUE_LENGTH) {
      sanitized = sanitized.substring(0, this.MAX_VALUE_LENGTH) + "...";
    }

    // Escape shell metacharacters for basic security
    sanitized = sanitized.replace(/([\\$`])/g, "\\$1");

    return sanitized;
  }

  async listApps(): Promise<string[]> {
    this.debug("Listing applications");

    const result = await this.executeCommand("list apps");
    if (!result.success) {
      this.warn("Failed to list apps", { error: result.error });
      return []; // Return empty array instead of throwing for menu bar
    }

    // Parse the output to extract app names
    const lines = result.output.split("\n");
    const apps: string[] = [];

    for (const line of lines) {
      const trimmed = line.trim();

      if (trimmed && trimmed.includes("•")) {
        // Extract app name from format like "  • ghostty"
        const match = trimmed.match(/•\s*(\w+)/);
        if (match) {
          apps.push(match[1]);
        }
      }
    }

    this.info(`Found ${apps.length} applications`, { apps });
    return apps;
  }

  async listValues(app: string): Promise<{ key: string; value: string }[]> {
    this.validateAppName(app);
    this.debug(`Listing configuration values for ${app}`);

    const result = await this.executeCommand("list values", [app]);
    if (!result.success) {
      this.recordError({
        operation: "listValues",
        app,
        timestamp: new Date(),
      });
      this.error(`Failed to list values for ${app}`, { error: result.error });
      throw new Error(result.error || `Failed to list values for ${app}`);
    }

    const values: { key: string; value: string }[] = [];
    const lines = result.output.split("\n");

    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed && trimmed.includes(":")) {
        const [key, ...valueParts] = trimmed.split(":");
        const value = valueParts.join(":").trim();
        if (key && value) {
          values.push({
            key: key.trim(),
            value: this.sanitizeValue(value),
          });
        }
      }
    }

    this.info(`Found ${values.length} configuration values for ${app}`);
    return values;
  }

  async listChanged(app: string): Promise<{ key: string; value: string; default: string }[]> {
    this.validateAppName(app);

    const result = await this.executeCommand("list changed", [app]);
    if (!result.success) {
      console.warn(`Failed to list changed values for ${app}:`, result.error);
      this.recordError({
        operation: "listChanged",
        app,
        timestamp: new Date(),
      });
      return []; // Return empty array for menu bar compatibility
    }

    const values: { key: string; value: string; default: string }[] = [];
    const lines = result.output.split("\n");

    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed && trimmed.includes(":")) {
        const [key, valuePart] = trimmed.split(":");
        if (valuePart && valuePart.includes("(default:")) {
          const valueMatch = valuePart.match(/^(.*?)\s*\(default:\s*(.*?)\)$/);
          if (valueMatch) {
            values.push({
              key: key.trim(),
              value: this.sanitizeValue(valueMatch[1].trim()),
              default: this.sanitizeValue(valueMatch[2].trim()),
            });
          }
        }
      }
    }

    return values;
  }

  async toggleConfig(app: string, key: string, value: string): Promise<string> {
    this.validateAppName(app);
    this.validateConfigKey(key);

    const sanitizedValue = this.sanitizeValue(value);

    const result = await this.executeCommand("toggle", [app, key, sanitizedValue]);
    if (!result.success) {
      this.recordError({
        operation: "toggleConfig",
        app,
        key,
        value: sanitizedValue,
        timestamp: new Date(),
      });
      throw new Error(result.error || `Failed to toggle ${key} for ${app}`);
    }
    return result.output;
  }

  async listKeymaps(app: string): Promise<{ keybind: string; action: string }[]> {
    this.validateAppName(app);

    const result = await this.executeCommand("keymap list", [app]);
    if (!result.success) {
      this.recordError({
        operation: "listKeymaps",
        app,
        timestamp: new Date(),
      });
      throw new Error(result.error || `Failed to list keymaps for ${app}`);
    }

    const keymaps: { keybind: string; action: string }[] = [];
    const lines = result.output.split("\n");

    for (const line of lines) {
      const trimmed = line.trim();
      if (trimmed && trimmed.includes("→")) {
        const [keybind, action] = trimmed.split("→").map((s) => s.trim());
        if (keybind && action) {
          keymaps.push({
            keybind: keybind.substring(0, 50), // Limit keybind length
            action: this.sanitizeValue(action),
          });
        }
      }
    }

    return keymaps;
  }
}

// Global instance
export const zeroui = new ZeroUI();
