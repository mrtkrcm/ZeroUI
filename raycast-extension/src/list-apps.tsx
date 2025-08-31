import { Action, ActionPanel, Icon, List, open, showToast, Toast } from "@raycast/api";
import { useCallback, useEffect, useState } from "react";
import { LogLevel, zeroui } from "./utils";

export default function ListAppsCommand() {
  const [apps, setApps] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const [lastRefresh, setLastRefresh] = useState<Date | null>(null);

  useEffect(() => {
    loadApps();
  }, []);

  const loadApps = useCallback(
    async (showSuccessToast = false) => {
      try {
        setIsLoading(true);
        setError(null);

        const appList = await zeroui.listApps();
        setApps(appList);
        setLastRefresh(new Date());
        setRetryCount(0); // Reset retry count on success

        if (showSuccessToast) {
          await showToast({
            style: Toast.Style.Success,
            title: "Apps Refreshed",
            message: `Found ${appList.length} applications`,
          });
        }
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : "Failed to load applications";
        setError(errorMessage);
        setRetryCount((prev) => prev + 1);

        // Show different messages based on retry count
        if (retryCount < 2) {
          await showToast({
            style: Toast.Style.Failure,
            title: "Failed to Load Apps",
            message: `${errorMessage}. Tap refresh to retry.`,
          });
        } else {
          await showToast({
            style: Toast.Style.Failure,
            title: "Multiple Load Failures",
            message: "Please check ZeroUI installation and try again.",
          });
        }

        console.error("Failed to load apps:", err);
      } finally {
        setIsLoading(false);
      }
    },
    [retryCount],
  );

  if (error) {
    return (
      <List>
        <List.Item
          title="Error Loading Applications"
          subtitle={error}
          icon={Icon.ExclamationMark}
          accessories={retryCount > 0 ? [{ text: `${retryCount} retries` }] : undefined}
          actions={
            <ActionPanel>
              <ActionPanel.Section>
                <Action
                  title="Retry"
                  icon={Icon.ArrowClockwise}
                  onAction={() => loadApps()}
                  shortcut={{ modifiers: ["cmd"], key: "r" }}
                />
                <Action
                  title="Clear Cache & Retry"
                  icon={Icon.Trash}
                  onAction={async () => {
                    zeroui.clearCache();
                    zeroui.clearErrorHistory();
                    await loadApps();
                  }}
                  shortcut={{ modifiers: ["cmd", "shift"], key: "r" }}
                />
              </ActionPanel.Section>
              <ActionPanel.Section>
                <Action
                  title="View Debug Logs"
                  icon={Icon.Info}
                  onAction={() => {
                    const stats = zeroui.getCacheStats();
                    const errors = zeroui.getErrorHistory();
                    const logs = zeroui.getLogs(LogLevel.WARN);
                    console.log("=== ZeroUI Debug Information ===");
                    console.log("Cache Stats:", stats);
                    console.log("Recent Errors:", errors);
                    console.log("Warning+ Logs:", logs);
                    console.log("=================================");
                    showToast({
                      style: Toast.Style.Animated,
                      title: "Debug Info Logged",
                      message: "Check console for detailed diagnostics",
                    });
                  }}
                />
              </ActionPanel.Section>
            </ActionPanel>
          }
        />
      </List>
    );
  }

  // Empty state handling
  if (!isLoading && apps.length === 0 && !error) {
    return (
      <List>
        <List.Item
          title="No Applications Found"
          subtitle="ZeroUI couldn't find any configured applications"
          icon={Icon.QuestionMark}
          actions={
            <ActionPanel>
              <ActionPanel.Section>
                <Action
                  title="Refresh"
                  icon={Icon.ArrowClockwise}
                  onAction={() => loadApps(true)}
                  shortcut={{ modifiers: ["cmd"], key: "r" }}
                />
                <Action
                  title="Check Zeroui Status"
                  icon={Icon.Info}
                  onAction={() => {
                    console.log("ZeroUI Cache Stats:", zeroui.getCacheStats());
                    showToast({
                      style: Toast.Style.Animated,
                      title: "Status Check",
                      message: "Cache stats logged to console",
                    });
                  }}
                />
              </ActionPanel.Section>
            </ActionPanel>
          }
        />
      </List>
    );
  }

  return (
    <List
      isLoading={isLoading}
      searchBarPlaceholder="Search applications..."
      filtering={true}
      throttle={true}
    >
      <List.Section
        title="Available Applications"
        subtitle={`${apps.length} apps${lastRefresh ? ` • Updated ${lastRefresh.toLocaleTimeString()}` : ""}`}
      >
        {apps.map((app) => (
          <List.Item
            key={app}
            title={app}
            subtitle="UI Configuration Manager"
            icon={Icon.AppWindow}
            keywords={[app.toLowerCase(), "configuration", "settings"]}
            actions={
              <ActionPanel>
                <ActionPanel.Section>
                  <Action.Push
                    title="View Configuration"
                    target={<AppConfigView app={app} />}
                    icon={Icon.Gear}
                    shortcut={{ modifiers: ["cmd"], key: "c" }}
                  />
                  <Action.Push
                    title="View Changed Values"
                    target={<AppChangedView app={app} />}
                    icon={Icon.CheckCircle}
                    shortcut={{ modifiers: ["cmd"], key: "d" }}
                  />
                  <Action.Push
                    title="View Keymaps"
                    target={<AppKeymapsView app={app} />}
                    icon={Icon.Keyboard}
                    shortcut={{ modifiers: ["cmd"], key: "k" }}
                  />
                  <Action.Push
                    title="Toggle Configuration"
                    target={<AppToggleView app={app} />}
                    icon={Icon.Gear}
                    shortcut={{ modifiers: ["cmd"], key: "t" }}
                  />
                </ActionPanel.Section>

                <ActionPanel.Section>
                  <Action
                    title="Refresh"
                    icon={Icon.ArrowClockwise}
                    onAction={async () => {
                      zeroui.clearCache();
                      await loadApps(true);
                    }}
                    shortcut={{ modifiers: ["cmd"], key: "r" }}
                  />
                  <Action.CopyToClipboard
                    title="Copy App Name"
                    content={app}
                    shortcut={{ modifiers: ["cmd"], key: "." }}
                  />
                </ActionPanel.Section>
              </ActionPanel>
            }
          />
        ))}
      </List.Section>
    </List>
  );
}

function AppConfigView({ app }: { app: string }) {
  const [values, setValues] = useState<{ key: string; value: string }[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadValues() {
      try {
        setIsLoading(true);
        const configValues = await zeroui.listValues(app);
        setValues(configValues);
      } catch (err) {
        await showToast({
          style: Toast.Style.Failure,
          title: "Failed to Load Config",
          message: err instanceof Error ? err.message : "Unknown error",
        });
      } finally {
        setIsLoading(false);
      }
    }

    loadValues();
  }, [app]);

  return (
    <List isLoading={isLoading} searchBarPlaceholder="Search configuration...">
      <List.Section title={`${app} Configuration`} subtitle={`${values.length} settings`}>
        {values.map((item) => (
          <List.Item
            key={item.key}
            title={item.key}
            subtitle={item.value}
            icon={Icon.Dot}
            actions={
              <ActionPanel>
                <Action.CopyToClipboard title="Copy Key" content={item.key} icon={Icon.Clipboard} />
                <Action.CopyToClipboard
                  title="Copy Value"
                  content={item.value}
                  icon={Icon.Clipboard}
                />
              </ActionPanel>
            }
          />
        ))}
      </List.Section>
    </List>
  );
}

function AppChangedView({ app }: { app: string }) {
  const [values, setValues] = useState<{ key: string; value: string; default: string }[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadChanged() {
      try {
        setIsLoading(true);
        const changedValues = await zeroui.listChanged(app);
        setValues(changedValues);
      } catch (err) {
        await showToast({
          style: Toast.Style.Failure,
          title: "Failed to Load Changes",
          message: err instanceof Error ? err.message : "Unknown error",
        });
      } finally {
        setIsLoading(false);
      }
    }

    loadChanged();
  }, [app]);

  return (
    <List isLoading={isLoading} searchBarPlaceholder="Search changed values...">
      <List.Section title={`${app} Changed Values`} subtitle={`${values.length} modified`}>
        {values.map((item) => (
          <List.Item
            key={item.key}
            title={item.key}
            subtitle={`Current: ${item.value}`}
            accessories={[{ text: `Default: ${item.default}`, icon: Icon.Info }]}
            icon={Icon.CheckCircle}
            actions={
              <ActionPanel>
                <Action.CopyToClipboard
                  title="Copy Current Value"
                  content={item.value}
                  icon={Icon.Clipboard}
                />
                <Action.CopyToClipboard
                  title="Copy Default Value"
                  content={item.default}
                  icon={Icon.Clipboard}
                />
              </ActionPanel>
            }
          />
        ))}
      </List.Section>
    </List>
  );
}

function AppKeymapsView({ app }: { app: string }) {
  const [keymaps, setKeymaps] = useState<{ keybind: string; action: string }[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadKeymaps() {
      try {
        setIsLoading(true);
        const keymapList = await zeroui.listKeymaps(app);
        setKeymaps(keymapList);
      } catch (err) {
        await showToast({
          style: Toast.Style.Failure,
          title: "Failed to Load Keymaps",
          message: err instanceof Error ? err.message : "Unknown error",
        });
      } finally {
        setIsLoading(false);
      }
    }

    loadKeymaps();
  }, [app]);

  return (
    <List isLoading={isLoading} searchBarPlaceholder="Search keymaps..." filtering={true}>
      <List.Section title={`${app} Keymaps`} subtitle={`${keymaps.length} shortcuts`}>
        {keymaps.map((item, index) => (
          <List.Item
            key={index}
            title={item.keybind}
            subtitle={item.action}
            icon={Icon.Keyboard}
            keywords={[item.keybind.toLowerCase(), item.action.toLowerCase()]}
            actions={
              <ActionPanel>
                <Action.CopyToClipboard
                  title="Copy Keybind"
                  content={item.keybind}
                  icon={Icon.Clipboard}
                  shortcut={{ modifiers: ["cmd"], key: "c" }}
                />
                <Action.CopyToClipboard
                  title="Copy Action"
                  content={item.action}
                  icon={Icon.Clipboard}
                  shortcut={{ modifiers: ["cmd"], key: "a" }}
                />
              </ActionPanel>
            }
          />
        ))}
      </List.Section>
    </List>
  );
}

function AppToggleView({ app }: { app: string }) {
  const [values, setValues] = useState<{ key: string; value: string }[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadValues() {
      try {
        setIsLoading(true);
        const configValues = await zeroui.listValues(app);
        setValues(configValues);
      } catch (err) {
        await showToast({
          style: Toast.Style.Failure,
          title: "Failed to Load Config",
          message: err instanceof Error ? err.message : "Unknown error",
        });
      } finally {
        setIsLoading(false);
      }
    }

    loadValues();
  }, [app]);

  const handleToggle = async (key: string, currentValue: string) => {
    try {
      await showToast({
        style: Toast.Style.Animated,
        title: "Toggling Configuration",
        message: `${key}: ${currentValue}`,
      });

      // Navigate to toggle-config command with pre-filled values
      const url = `raycast://extensions/zeroui/zeroui/toggle-config?app=${encodeURIComponent(app)}&key=${encodeURIComponent(key)}&value=${encodeURIComponent(currentValue)}`;
      await open(url);
    } catch (err) {
      await showToast({
        style: Toast.Style.Failure,
        title: "Failed to Toggle",
        message: err instanceof Error ? err.message : "Unknown error",
      });
    }
  };

  return (
    <List
      isLoading={isLoading}
      searchBarPlaceholder="Search configuration values..."
      filtering={true}
    >
      <List.Section title={`${app} Configuration`} subtitle={`${values.length} settings`}>
        {values.map((item) => (
          <List.Item
            key={item.key}
            title={item.key}
            subtitle={item.value}
            icon={Icon.Dot}
            keywords={[item.key.toLowerCase(), item.value.toLowerCase()]}
            actions={
              <ActionPanel>
                <Action
                  title="Toggle Value"
                  icon={Icon.ArrowClockwise}
                  onAction={() => handleToggle(item.key, item.value)}
                  shortcut={{ modifiers: ["cmd"], key: "t" }}
                />
                <Action.CopyToClipboard
                  title="Copy Key"
                  content={item.key}
                  icon={Icon.Clipboard}
                  shortcut={{ modifiers: ["cmd"], key: "c" }}
                />
                <Action.CopyToClipboard
                  title="Copy Value"
                  content={item.value}
                  icon={Icon.Clipboard}
                  shortcut={{ modifiers: ["cmd"], key: "v" }}
                />
              </ActionPanel>
            }
          />
        ))}
      </List.Section>
    </List>
  );
}
