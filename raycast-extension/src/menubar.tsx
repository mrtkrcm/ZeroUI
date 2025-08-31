import { Icon, MenuBarExtra, open } from "@raycast/api";
import { useEffect, useState } from "react";
import { zeroui } from "./utils";

interface AppInfo {
  name: string;
  status: "configured" | "default" | "error";
  changedCount: number;
}

export default function MenuBarCommand() {
  const [apps, setApps] = useState<AppInfo[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());
  const [errorCount, setErrorCount] = useState(0);
  const [lastError, setLastError] = useState<string | null>(null);

  const fetchApps = async () => {
    try {
      setIsLoading(true);
      setLastError(null);

      const appList = await zeroui.listApps();

      const appsWithStatus = await Promise.all(
        appList.map(async (appName) => {
          try {
            const changedValues = await zeroui.listChanged(appName);
            return {
              name: appName,
              status: changedValues.length > 0 ? "configured" : "default",
              changedCount: changedValues.length,
            } as AppInfo;
          } catch (error) {
            console.warn(`Failed to fetch status for ${appName}:`, error);
            setErrorCount((prev) => prev + 1);
            return {
              name: appName,
              status: "error" as const,
              changedCount: 0,
            } as AppInfo;
          }
        }),
      );

      setApps(appsWithStatus);
      setLastUpdate(new Date());
      setErrorCount(0); // Reset error count on successful fetch
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : "Failed to fetch applications";
      console.error("Failed to fetch apps:", error);
      setLastError(errorMessage);
      setApps([]);
      setErrorCount((prev) => prev + 1);
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    fetchApps();
  }, []);

  const getMenuBarTitle = () => {
    if (isLoading) return "⏳";

    const configuredCount = apps.filter((app) => app.status === "configured").length;
    const totalCount = apps.length;

    if (totalCount === 0) return "🔧";
    if (configuredCount === 0) return `🔧 ${totalCount}`;
    return `🔧 ${configuredCount}/${totalCount}`;
  };

  const getMenuBarIcon = () => {
    if (isLoading) return Icon.Clock;
    if (lastError) return Icon.ExclamationMark;

    const hasConfigured = apps.some((app) => app.status === "configured");
    const hasErrors = apps.some((app) => app.status === "error");

    if (hasErrors && errorCount > 0) return Icon.ExclamationMark;
    if (hasConfigured) return Icon.Gear;
    return Icon.Cog;
  };

  const openZeroUICommand = (command: string) => {
    const url = `raycast://extensions/zeroui/zeroui/${command}`;
    open(url);
  };

  return (
    <MenuBarExtra
      title={getMenuBarTitle()}
      icon={getMenuBarIcon()}
      tooltip={`ZeroUI - ${apps.length} apps${lastError ? ` • Error: ${lastError}` : ""} (Updated: ${lastUpdate.toLocaleTimeString()})`}
    >
      {/* Header */}
      <MenuBarExtra.Item
        title="ZeroUI Configuration Manager"
        icon={Icon.Gear}
        onAction={() => openZeroUICommand("list-apps")}
      />

      <MenuBarExtra.Separator />

      {/* Applications Section */}
      {apps.length > 0 ? (
        <>
          <MenuBarExtra.Item title="Applications" />
          {apps.map((app) => (
            <MenuBarExtra.Submenu
              key={app.name}
              title={app.name}
              icon={
                app.status === "configured"
                  ? Icon.CheckCircle
                  : app.status === "error"
                    ? Icon.XMarkCircle
                    : Icon.Circle
              }
            >
              <MenuBarExtra.Item
                title="View Configuration"
                icon={Icon.Document}
                onAction={() => openZeroUICommand("list-values")}
              />

              {app.status === "configured" && (
                <MenuBarExtra.Item
                  title={`Changed Values (${app.changedCount})`}
                  icon={Icon.ExclamationMark}
                  onAction={() => openZeroUICommand("list-changed")}
                />
              )}

              <MenuBarExtra.Item
                title="View Keymaps"
                icon={Icon.Keyboard}
                onAction={() => openZeroUICommand("keymap-list")}
              />

              <MenuBarExtra.Separator />

              <MenuBarExtra.Item
                title="Toggle Setting..."
                icon={Icon.Gear}
                onAction={() => openZeroUICommand("toggle-config")}
              />
            </MenuBarExtra.Submenu>
          ))}
        </>
      ) : (
        <MenuBarExtra.Item
          title={isLoading ? "Loading applications..." : "No applications found"}
          icon={isLoading ? Icon.Clock : Icon.ExclamationMark}
        />
      )}

      <MenuBarExtra.Separator />

      {/* Quick Actions */}
      <MenuBarExtra.Item title="Quick Actions" />
      <MenuBarExtra.Item
        title="List All Applications"
        icon={Icon.List}
        onAction={() => openZeroUICommand("list-apps")}
      />
      <MenuBarExtra.Item
        title="Manage Presets"
        icon={Icon.Tray}
        onAction={() => openZeroUICommand("manage-presets")}
      />

      <MenuBarExtra.Separator />

      {/* Refresh */}
      <MenuBarExtra.Item
        title="Refresh & Clear Cache"
        icon={Icon.ArrowClockwise}
        onAction={() => {
          zeroui.clearCache();
          fetchApps();
        }}
        shortcut={{ modifiers: ["cmd"], key: "r" }}
      />

      {/* Cache Stats */}
      <MenuBarExtra.Item title={`Cache: ${zeroui.getCacheStats().size} items`} icon={Icon.Info} />

      {/* Status */}
      <MenuBarExtra.Item
        title={`Last updated: ${lastUpdate.toLocaleTimeString()}`}
        icon={Icon.Clock}
      />
    </MenuBarExtra>
  );
}
