import { Action, ActionPanel, Icon, List } from "@raycast/api";
import { usePromise } from "@raycast/utils";
import { useEffect, useState } from "react";
import { zeroui } from "./utils";

interface ListChangedProps {
  arguments?: {
    app?: string;
  };
}

export default function ListChangedCommand(props: ListChangedProps) {
  const { arguments: args } = props;
  const [selectedApp, setSelectedApp] = useState<string>(args?.app || "");
  const [apps, setApps] = useState<string[]>([]);

  // Load apps
  const { data: appList, isLoading: isLoadingApps } = usePromise(async () => {
    const list = await zeroui.listApps();
    setApps(list);
    return list;
  }, []);

  // Load changed values if app is selected
  const { data: values, isLoading: isLoadingValues } = usePromise(
    async (app: string) => {
      if (!app) return [];
      return await zeroui.listChanged(app);
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

  const isLoading = isLoadingApps || isLoadingValues;

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
                    title="View Changed Values"
                    onAction={() => {
                      setSelectedApp(app);
                      // values will auto-load via usePromise dependency
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
    <List searchBarPlaceholder="Search changed values...">
      <List.Section
        title={`${selectedApp} Changed Values`}
        subtitle={`${values?.length || 0} modified settings`}
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
                  title="Copy Current"
                  content={item.value}
                />
                <Action.CopyToClipboard
                  title="Copy Default"
                  content={item.default}
                />
                <Action
                  title="Back to Apps"
                  icon={Icon.ArrowLeft}
                  onAction={() => {
                    setSelectedApp("");
                    // values cleared via dependency change
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
