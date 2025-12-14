import { Action, ActionPanel, Icon, List } from "@raycast/api";
import { usePromise } from "@raycast/utils";
import { useEffect, useState } from "react";
import { zeroui } from "./utils";

interface KeymapListProps {
  arguments?: {
    app?: string;
  };
}

export default function KeymapListCommand(props: KeymapListProps) {
  const { arguments: args } = props;
  const [selectedApp, setSelectedApp] = useState<string>(args?.app || "");
  const [apps, setApps] = useState<string[]>([]);

  // Load apps
  const { data: appList, isLoading: isLoadingApps } = usePromise(async () => {
    const list = await zeroui.listApps();
    setApps(list);
    return list;
  }, []);

  // Load keymaps if app is selected
  const { data: keymaps, isLoading: isLoadingKeymaps } = usePromise(
    async (app: string) => {
      if (!app) return [];
      return await zeroui.listKeymaps(app);
    },
    [selectedApp],
    {
      execute: !!selectedApp,
    },
  );

  useEffect(() => {
    if (args?.app && appList?.includes(args.app)) {
      setSelectedApp(args.app);
    }
  }, [args, appList]);

  const isLoading = isLoadingApps || isLoadingKeymaps;

  if (!selectedApp) {
    return (
      <List isLoading={isLoading} searchBarPlaceholder="Search applications...">
        <List.Section title="Select Application">
          {(apps || []).map((app) => (
            <List.Item
              key={app}
              title={app}
              icon={Icon.AppWindow}
              actions={
                <ActionPanel>
                  <Action
                    title="View Keymaps"
                    onAction={() => {
                      setSelectedApp(app);
                      // keymaps will auto-load via usePromise dependency
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
      <List.Section
        title={`${selectedApp} Keymaps`}
        subtitle={`${keymaps?.length || 0} shortcuts`}
      >
        {(keymaps || []).map((item, index) => (
          <List.Item
            key={index}
            title={item.keybind}
            subtitle={item.action}
            icon={Icon.Keyboard}
            actions={
              <ActionPanel>
                <Action.CopyToClipboard
                  title="Copy Keybind"
                  content={item.keybind}
                />
                <Action.CopyToClipboard
                  title="Copy Action"
                  content={item.action}
                />
                <Action
                  title="Back to Apps"
                  icon={Icon.ArrowLeft}
                  onAction={() => {
                    setSelectedApp("");
                    // keymaps cleared via dependency change
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
