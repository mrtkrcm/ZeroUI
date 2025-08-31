import { Action, ActionPanel, Form, Icon, List, showToast, Toast } from "@raycast/api";
import { useEffect, useState } from "react";
import { zeroui } from "./utils";

interface ToggleConfigProps {
  arguments?: {
    app?: string;
    key?: string;
    value?: string;
  };
}

export default function ToggleConfigCommand(props: ToggleConfigProps) {
  const { arguments: args } = props;
  const [apps, setApps] = useState<string[]>([]);
  const [selectedApp, setSelectedApp] = useState<string>(args?.app || "");
  const [configValues, setConfigValues] = useState<{ key: string; value: string }[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    async function loadApps() {
      try {
        setIsLoading(true);
        const appList = await zeroui.listApps();
        setApps(appList);

        // If app was provided in arguments, load its config
        if (args?.app && appList.includes(args.app)) {
          setSelectedApp(args.app);
          const values = await zeroui.listValues(args.app);
          setConfigValues(values);
        }
      } catch (err) {
        await showToast({
          style: Toast.Style.Failure,
          title: "Failed to Load Apps",
          message: err instanceof Error ? err.message : "Unknown error",
        });
      } finally {
        setIsLoading(false);
      }
    }

    loadApps();
  }, [args]);

  // If we have all arguments, show the toggle form directly
  if (args?.app && args?.key && args?.value) {
    return <ToggleForm app={args.app} key={args.key} value={args.value} />;
  }

  // If we have an app selected, show the config values
  if (selectedApp && configValues.length > 0) {
    return (
      <List searchBarPlaceholder="Search configuration to toggle...">
        <List.Section title={`${selectedApp} Configuration`} subtitle="Click to toggle">
          {configValues.map((item) => (
            <List.Item
              key={item.key}
              title={item.key}
              subtitle={`Current: ${item.value}`}
              icon={Icon.Dot}
              actions={
                <ActionPanel>
                  <Action.Push
                    title="Toggle Value"
                    target={<ToggleForm app={selectedApp} key={item.key} value={item.value} />}
                    icon={Icon.Switch}
                  />
                  <Action.CopyToClipboard
                    title="Copy Current Value"
                    content={item.value}
                    icon={Icon.Clipboard}
                  />
                  <Action
                    title="Back to Apps"
                    icon={Icon.ArrowLeft}
                    onAction={() => {
                      setSelectedApp("");
                      setConfigValues([]);
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

  // Show app selection
  return (
    <List isLoading={isLoading} searchBarPlaceholder="Search applications...">
      <List.Section title="Select Application" subtitle="Choose app to configure">
        {apps.map((app) => (
          <List.Item
            key={app}
            title={app}
            subtitle="Click to view configuration"
            icon={Icon.AppWindow}
            actions={
              <ActionPanel>
                <Action
                  title="Select App"
                  icon={Icon.CheckCircle}
                  onAction={async () => {
                    try {
                      setSelectedApp(app);
                      const values = await zeroui.listValues(app);
                      setConfigValues(values);
                    } catch (err) {
                      await showToast({
                        style: Toast.Style.Failure,
                        title: "Failed to Load Config",
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

function ToggleForm({ app, key, value }: { app: string; key: string; value: string }) {
  const [newValue, setNewValue] = useState(value);

  const handleSubmit = async (values: { newValue: string }) => {
    if (!values.newValue.trim()) {
      await showToast({
        style: Toast.Style.Failure,
        title: "Invalid Value",
        message: "Please enter a value to toggle to",
      });
      return;
    }

    try {
      await showToast({
        style: Toast.Style.Animated,
        title: "Toggling Configuration...",
        message: `${key}: ${value} → ${values.newValue}`,
      });

      await zeroui.toggleConfig(app, key, values.newValue);

      await showToast({
        style: Toast.Style.Success,
        title: "Configuration Toggled",
        message: `${key} updated successfully`,
      });

      // Close the form after successful toggle
      setTimeout(() => {
        // This will close the Raycast window
        process.exit(0);
      }, 1000);
    } catch (err) {
      await showToast({
        style: Toast.Style.Failure,
        title: "Toggle Failed",
        message: err instanceof Error ? err.message : "Unknown error occurred",
      });
    }
  };

  return (
    <Form
      actions={
        <ActionPanel>
          <Action.SubmitForm
            title="Toggle Configuration"
            icon={Icon.CheckCircle}
            onSubmit={handleSubmit}
          />
        </ActionPanel>
      }
    >
      <Form.Description
        title="Toggle Configuration"
        text={`Application: ${app}\nKey: ${key}\nCurrent Value: ${value}`}
      />
      <Form.TextField
        id="newValue"
        title="New Value"
        placeholder="Enter new value"
        value={newValue}
        onChange={setNewValue}
        storeValue={true}
      />
      <Form.Description text="Enter the new value for this configuration setting." />
    </Form>
  );
}
