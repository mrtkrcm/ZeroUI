import { Action, ActionPanel, Icon, List, showToast, Toast } from "@raycast/api";
import { useEffect, useState } from "react";
import { zeroui } from "./utils";

interface ListChangedProps {
  arguments?: {
    app?: string;
  };
}

export default function ListChangedCommand(props: ListChangedProps) {
  const { arguments: args } = props;
  const [values, setValues] = useState<{ key: string; value: string; default: string }[]>([]);
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
          const changedValues = await zeroui.listChanged(args.app);
          setValues(changedValues);
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
                    title="View Changed Values"
                    onAction={async () => {
                      try {
                        const changedValues = await zeroui.listChanged(app);
                        setValues(changedValues);
                        setSelectedApp(app);
                      } catch (err) {
                        await showToast({
                          style: Toast.Style.Failure,
                          title: "Failed to Load Changed Values",
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
    <List searchBarPlaceholder="Search changed values...">
      <List.Section
        title={`${selectedApp} Changed Values`}
        subtitle={`${values.length} modified settings`}
      >
        {values.map((item) => (
          <List.Item
            key={item.key}
            title={item.key}
            subtitle={`Current: ${item.value}`}
            accessories={[{ text: `Default: ${item.default}`, icon: Icon.Info }]}
            icon={Icon.CheckCircle}
            actions={
              <ActionPanel>
                <Action.CopyToClipboard title="Copy Current" content={item.value} />
                <Action.CopyToClipboard title="Copy Default" content={item.default} />
                <Action
                  title="Back to Apps"
                  icon={Icon.ArrowLeft}
                  onAction={() => {
                    setSelectedApp("");
                    setValues([]);
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
