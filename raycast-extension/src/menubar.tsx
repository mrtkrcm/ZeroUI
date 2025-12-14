import { Icon, MenuBarExtra, open } from "@raycast/api";
import { useCachedPromise } from "@raycast/utils";
import { zeroui } from "./utils";

interface AppInfo {
  name: string;
  status: "configured" | "default" | "error";
  changedCount: number;
}

export default function MenuBarCommand() {
  const {
    data: apps,
    isLoading,
    error,
    revalidate,
  } = useCachedPromise(
    async () => {
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
            return {
              name: appName,
              status: "error" as const,
              changedCount: 0,
            } as AppInfo;
          }
        }),
      );

      return appsWithStatus;
    },
    [],
    {
      initialData: [],
      keepPreviousData: true,
    },
  );

  const getMenuBarTitle = () => {
    if (isLoading && (!apps || apps.length === 0)) return "⏳";

    const configuredCount = (apps || []).filter(
      (app) => app.status === "configured",
    ).length;
    const totalCount = apps?.length || 0;

    if (totalCount === 0) return "🔧";
    if (configuredCount === 0) return `🔧 ${totalCount}`;
    return `🔧 ${configuredCount}/${totalCount}`;
  };

  const getMenuBarIcon = () => {
    if (isLoading && (!apps || apps.length === 0)) return Icon.Clock;
    if (error) return Icon.ExclamationMark;

    const hasConfigured = (apps || []).some(
      (app) => app.status === "configured",
    );
    const hasErrors = (apps || []).some((app) => app.status === "error");

    if (hasErrors) return Icon.ExclamationMark;
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
      tooltip={`ZeroUI - ${apps?.length || 0} apps${error ? ` • Error: ${error.message}` : ""}`}
    >
      {/* Header */}
      <MenuBarExtra.Item
        title="ZeroUI Configuration Manager"
        icon={Icon.Gear}
        onAction={() => openZeroUICommand("list-apps")}
      />

      <MenuBarExtra.Separator />

      {/* Applications Section */}
      {apps && apps.length > 0 ? (
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
          title={
            isLoading ? "Loading applications..." : "No applications found"
          }
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
          revalidate();
        }}
        shortcut={{ modifiers: ["cmd"], key: "r" }}
      />

      {/* Cache Stats */}
      <MenuBarExtra.Item
        title={`Cache: ${zeroui.getCacheStats().size} items`}
        icon={Icon.Info}
      />

      <MenuBarExtra.Item
        title={isLoading ? "Updating..." : "Updated just now"}
        icon={Icon.Clock}
      />
    </MenuBarExtra>
  );
}
