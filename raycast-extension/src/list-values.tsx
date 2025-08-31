import { Action, ActionPanel, Icon, List, showToast, Toast } from "@raycast/api";
import { useEffect, useState } from "react";
import { zeroui } from "./utils";

interface ListValuesProps {
  arguments?: {
    app?: string;
  };
}

export default function ListValuesCommand(props: ListValuesProps) {
  const { arguments: args } = props;
  const [values, setValues] = useState<{ key: string; value: string }[]>([]);
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
          const configValues = await zeroui.listValues(args.app);
          setValues(configValues);
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
                    title="View Values"
                    onAction={async () => {
                      try {
                        const configValues = await zeroui.listValues(app);
                        setValues(configValues);
                        setSelectedApp(app);
                      } catch (err) {
                        await showToast({
                          style: Toast.Style.Failure,
                          title: "Failed to Load Values",
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
    <List searchBarPlaceholder="Search configuration values...">
      <List.Section
        title={`${selectedApp} Configuration Values`}
        subtitle={`${values.length} settings`}
      >
        {values.map((item) => (
          <List.Item
            key={item.key}
            title={item.key}
            subtitle={item.value}
            icon={Icon.Dot}
            actions={
              <ActionPanel>
                <Action.CopyToClipboard title="Copy Key" content={item.key} />
                <Action.CopyToClipboard title="Copy Value" content={item.value} />
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
