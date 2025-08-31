import { Action, ActionPanel, Icon, List, showToast, Toast } from "@raycast/api";
import { useEffect, useState } from "react";
import { zeroui } from "./utils";

interface KeymapListProps {
  arguments?: {
    app?: string;
  };
}

export default function KeymapListCommand(props: KeymapListProps) {
  const { arguments: args } = props;
  const [keymaps, setKeymaps] = useState<{ keybind: string; action: string }[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [selectedApp, setSelectedApp] = useState<string>(args?.app || "");
  const [apps, setApps] = useState<string[]>([]);

  useEffect(() => {
    async function loadData() {
      try {
        setIsLoading(true);
        const appList = await zeroui.listApps();
        setApps(appList);

        if (args?.app && appList.includes(args.app)) {
          const keymapList = await zeroui.listKeymaps(args.app);
          setKeymaps(keymapList);
          setSelectedApp(args.app);
        }
      } catch (err) {
        await showToast({
          style: Toast.Style.Failure,
          title: "Failed to Load",
          message: err instanceof Error ? err.message : "Unknown error",
        });
      } finally {
        setIsLoading(false);
      }
    }

    loadData();
  }, [args]);

  if (!selectedApp) {
    return (
      <List isLoading={isLoading} searchBarPlaceholder="Search applications...">
        <List.Section title="Select Application">
          {apps.map((app) => (
            <List.Item
              key={app}
              title={app}
              icon={Icon.AppWindow}
              actions={
                <ActionPanel>
                  <Action
                    title="View Keymaps"
                    onAction={async () => {
                      try {
                        const keymapList = await zeroui.listKeymaps(app);
                        setKeymaps(keymapList);
                        setSelectedApp(app);
                      } catch (err) {
                        await showToast({
                          style: Toast.Style.Failure,
                          title: "Failed to Load Keymaps",
                          message: err instanceof Error ? err.message : "Unknown error",
                        });
                      }
                    }}
                  />
                </ActionPanel>
              }
            />
          ))}
        </List.Section>
      </List>
    );
  }

  return (
    <List searchBarPlaceholder="Search keymaps...">
      <List.Section title={`${selectedApp} Keymaps`} subtitle={`${keymaps.length} shortcuts`}>
        {keymaps.map((item, index) => (
          <List.Item
            key={index}
            title={item.keybind}
            subtitle={item.action}
            icon={Icon.Keyboard}
            actions={
              <ActionPanel>
                <Action.CopyToClipboard title="Copy Keybind" content={item.keybind} />
                <Action.CopyToClipboard title="Copy Action" content={item.action} />
                <Action
                  title="Back to Apps"
                  icon={Icon.ArrowLeft}
                  onAction={() => {
                    setSelectedApp("");
                    setKeymaps([]);
                  }}
                />
              </ActionPanel>
            }
          />
        ))}
      </List.Section>
    </List>
  );
}
