import {
  Action,
  ActionPanel,
  Form,
  Icon,
  List,
  showToast,
  Toast,
} from "@raycast/api";
import { useForm, usePromise } from "@raycast/utils";
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
  const [selectedApp, setSelectedApp] = useState<string>(args?.app || "");
  const [apps, setApps] = useState<string[]>([]);

  // Load apps
  const { data: appList, isLoading: isLoadingApps } = usePromise(async () => {
    const list = await zeroui.listApps();
    setApps(list);
    return list;
  }, []);

  // Load config values if app is selected
  const { data: configValues, isLoading: isLoadingConfig } = usePromise(
    async (app: string) => {
      if (!app) return [];
      return await zeroui.listValues(app);
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

  const isLoading = isLoadingApps || isLoadingConfig;

  // If we have all arguments, show the toggle form directly
  if (args?.app && args?.key && args?.value) {
    return (
      <ToggleForm app={args.app} configKey={args.key} value={args.value} />
    );
  }

  // If we have an app selected, show the config values
  if (selectedApp && configValues && configValues.length > 0) {
    return (
      <List
        isLoading={isLoading}
        searchBarPlaceholder="Search configuration to toggle..."
      >
        <List.Section
          title={`${selectedApp} Configuration`}
          subtitle="Click to toggle"
        >
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
                    target={
                      <ToggleForm
                        app={selectedApp}
                        configKey={item.key}
                        value={item.value}
                      />
                    }
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
      <List.Section
        title="Select Application"
        subtitle="Choose app to configure"
      >
        {(apps || []).map((app) => (
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
                  onAction={() => {
                    setSelectedApp(app);
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

function ToggleForm({
  app,
  configKey,
  value,
}: {
  app: string;
  configKey: string;
  value: string;
}) {
  const { handleSubmit, itemProps } = useForm<{ newValue: string }>({
    initialValues: {
      newValue: value,
    },
    validation: {
      newValue: (val) => (!val?.trim() ? "Value cannot be empty" : undefined),
    },
    onSubmit: async (values) => {
      try {
        await showToast({
          style: Toast.Style.Animated,
          title: "Toggling Configuration...",
          message: `${configKey}: ${value} → ${values.newValue}`,
        });

        await zeroui.toggleConfig(app, configKey, values.newValue);

        await showToast({
          style: Toast.Style.Success,
          title: "Configuration Toggled",
          message: `${configKey} updated successfully`,
        });

        setTimeout(() => {
          process.exit(0);
        }, 1000);
      } catch (err) {
        await showToast({
          style: Toast.Style.Failure,
          title: "Toggle Failed",
          message:
            err instanceof Error ? err.message : "Unknown error occurred",
        });
      }
    },
  });

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
        text={`Application: ${app}\nKey: ${configKey}\nCurrent Value: ${value}`}
      />
      <Form.TextField
        title="New Value"
        placeholder="Enter new value"
        {...itemProps.newValue}
      />
      <Form.Description text="Enter the new value for this configuration setting." />
    </Form>
  );
}
