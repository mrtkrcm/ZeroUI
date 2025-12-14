import {
  Action,
  ActionPanel,
  Icon,
  List,
  open,
  showToast,
  Toast,
} from "@raycast/api";
import { usePromise } from "@raycast/utils";
import { LogLevel, zeroui } from "./utils";

export default function ListAppsCommand() {
  const {
    data: apps,
    isLoading,
    error,
    revalidate,
  } = usePromise(
    async () => {
      const appList = await zeroui.listApps();
      return appList;
    },
    [],
    {
      onError: (error) => {
        showToast({
          style: Toast.Style.Failure,
          title: "Failed to Load Apps",
          message: error.message,
        });
      },
    },
  );

  if (error) {
    return (
      <List>
        <List.Item
          title="Error Loading Applications"
          subtitle={error.message}
          icon={Icon.ExclamationMark}
          actions={
            <ActionPanel>
              <ActionPanel.Section>
                <Action
                  title="Retry"
                  icon={Icon.ArrowClockwise}
                  onAction={() => revalidate()}
                  shortcut={{ modifiers: ["cmd"], key: "r" }}
                />
                <Action
                  title="Clear Cache & Retry"
                  icon={Icon.Trash}
                  onAction={async () => {
                    zeroui.clearCache();
                    zeroui.clearErrorHistory();
                    revalidate();
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
                  onAction={() => revalidate()}
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
      <List.EmptyView
        title="No Applications Found"
        description="ZeroUI found no configured applications. Ensure your apps are configured in your ZeroUI configuration directory."
        icon={Icon.MagnifyingGlass}
        actions={
          <ActionPanel>
            <Action
              title="Retry"
              icon={Icon.ArrowClockwise}
              onAction={() => revalidate()}
              shortcut={{ modifiers: ["cmd"], key: "r" }}
            />
          </ActionPanel>
        }
      />
      <List.Section
        title="Available Applications"
        subtitle={`${apps?.length || 0} apps`}
      >
        {(apps || []).map((app) => (
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
                      revalidate();
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
  const { data: values, isLoading } = usePromise(async () => {
    return await zeroui.listValues(app);
  }, [app]);

  return (
    <List isLoading={isLoading} searchBarPlaceholder="Search configuration...">
      <List.Section
        title={`${app} Configuration`}
        subtitle={`${values?.length || 0} settings`}
      >
        {(values || []).map((item) => (
          <List.Item
            key={item.key}
            title={item.key}
            subtitle={item.value}
            icon={Icon.Dot}
            actions={
              <ActionPanel>
                <Action.CopyToClipboard
                  title="Copy Key"
                  content={item.key}
                  icon={Icon.Clipboard}
                />
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
  const { data: values, isLoading } = usePromise(async () => {
    return await zeroui.listChanged(app);
  }, [app]);

  return (
    <List isLoading={isLoading} searchBarPlaceholder="Search changed values...">
      <List.Section
        title={`${app} Changed Values`}
        subtitle={`${values?.length || 0} modified`}
      >
        {(values || []).map((item) => (
          <List.Item
            key={item.key}
            title={item.key}
            subtitle={`Current: ${item.value}`}
            accessories={[
              { text: `Default: ${item.default}`, icon: Icon.Info },
            ]}
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
  const { data: keymaps, isLoading } = usePromise(async () => {
    return await zeroui.listKeymaps(app);
  }, [app]);

  return (
    <List
      isLoading={isLoading}
      searchBarPlaceholder="Search keymaps..."
      filtering={true}
    >
      <List.Section
        title={`${app} Keymaps`}
        subtitle={`${keymaps?.length || 0} shortcuts`}
      >
        {(keymaps || []).map((item, index) => (
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
  const { data: values, isLoading } = usePromise(async () => {
    return await zeroui.listValues(app);
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
      <List.Section
        title={`${app} Configuration`}
        subtitle={`${values?.length || 0} settings`}
      >
        {(values || []).map((item) => (
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
