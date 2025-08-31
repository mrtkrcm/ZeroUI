import { Action, ActionPanel, Icon, List, showToast, Toast } from "@raycast/api";
import { useEffect, useState } from "react";
import { zeroui } from "./utils";

interface CacheInfo {
  size: number;
  hits: number;
  misses: number;
  duration: number;
}

export default function CacheManagementCommand() {
  const [cacheInfo, setCacheInfo] = useState<CacheInfo | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadCacheInfo();
  }, []);

  const loadCacheInfo = () => {
    try {
      const stats = zeroui.getCacheStats();
      setCacheInfo({
        ...stats,
        duration: 0, // Cache doesn't track duration
      });
    } catch (error) {
      console.error("Failed to load cache info:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const clearCache = async () => {
    try {
      await showToast({
        style: Toast.Style.Animated,
        title: "Clearing Cache",
        message: "Removing all cached data...",
      });

      zeroui.clearCache();
      loadCacheInfo();

      await showToast({
        style: Toast.Style.Success,
        title: "Cache Cleared",
        message: "All cached data has been removed",
      });
    } catch (error) {
      await showToast({
        style: Toast.Style.Failure,
        title: "Failed to Clear Cache",
        message: error instanceof Error ? error.message : "Unknown error",
      });
    }
  };

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return "0 B";
    const k = 1024;
    const sizes = ["B", "KB", "MB", "GB"];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
  };

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  };

  return (
    <List isLoading={isLoading} searchBarPlaceholder="Cache management...">
      <List.Section title="Cache Information">
        {cacheInfo && (
          <>
            <List.Item
              title="Cache Size"
              subtitle={`${cacheInfo.size} entries`}
              icon={Icon.ArchiveBox}
              accessories={[{ text: formatBytes(cacheInfo.size * 100), icon: Icon.Info }]}
            />
            <List.Item
              title="Cache Hits"
              subtitle={`${cacheInfo.hits} successful lookups`}
              icon={Icon.Checkmark}
            />
            <List.Item
              title="Cache Misses"
              subtitle={`${cacheInfo.misses} cache misses`}
              icon={Icon.XMark}
            />
            <List.Item
              title="Performance"
              subtitle={`Avg response: ${formatDuration(cacheInfo.duration)}`}
              icon={Icon.Gauge}
            />
          </>
        )}
      </List.Section>

      <List.Section title="Actions">
        <List.Item
          title="Clear All Cache"
          subtitle="Remove all cached data to free memory"
          icon={Icon.Trash}
          actions={
            <ActionPanel>
              <Action
                title="Clear Cache"
                icon={Icon.Trash}
                onAction={clearCache}
                shortcut={{ modifiers: ["cmd"], key: "delete" }}
              />
              <Action
                title="Refresh Stats"
                icon={Icon.ArrowClockwise}
                onAction={loadCacheInfo}
                shortcut={{ modifiers: ["cmd"], key: "r" }}
              />
            </ActionPanel>
          }
        />
        <List.Item
          title="Enable/Disable Caching"
          subtitle="Toggle caching on/off in preferences"
          icon={Icon.Gear}
          actions={
            <ActionPanel>
              <Action.OpenExtensionPreferences title="Open Preferences" icon={Icon.Gear} />
            </ActionPanel>
          }
        />
      </List.Section>
    </List>
  );
}
