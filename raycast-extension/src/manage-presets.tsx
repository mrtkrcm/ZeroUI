import {
  Action,
  ActionPanel,
  Form,
  Icon,
  List,
  showToast,
  Toast,
} from "@raycast/api";
import { usePromise } from "@raycast/utils";
import { zeroui } from "./utils";

export default function ManagePresetsCommand() {
  const { data: apps, isLoading } = usePromise(async () => {
    return await zeroui.listApps();
  }, []);

  return (
    <List isLoading={isLoading} searchBarPlaceholder="Search applications...">
      <List.Section
        title="Manage Configuration Presets"
        subtitle="Apply predefined configurations"
      >
        {(apps || []).map((app) => (
          <List.Item
            key={app}
            title={app}
            subtitle="Apply preset configuration"
            icon={Icon.AppWindow}
            actions={
              <ActionPanel>
                <Action.Push
                  title="Apply Minimal Preset"
                  target={<ApplyPresetForm app={app} preset="minimal" />}
                  icon={Icon.Minus}
                />
                <Action.Push
                  title="Apply Default Preset"
                  target={<ApplyPresetForm app={app} preset="default" />}
                  icon={Icon.Circle}
                />
                <Action.Push
                  title="Apply Developer Preset"
                  target={<ApplyPresetForm app={app} preset="developer" />}
                  icon={Icon.Code}
                />
              </ActionPanel>
            }
          />
        ))}
      </List.Section>
    </List>
  );
}

function ApplyPresetForm({ app, preset }: { app: string; preset: string }) {
  const handleSubmit = async () => {
    try {
      await showToast({
        style: Toast.Style.Animated,
        title: "Applying Preset...",
        message: `Applying ${preset} preset to ${app}`,
      });

      // Use the preset command
      const result = await zeroui.executeCommand("preset", [app, preset]);

      if (result.success) {
        await showToast({
          style: Toast.Style.Success,
          title: "Preset Applied",
          message: `${preset} preset applied to ${app} successfully`,
        });
      } else {
        throw new Error(result.error || "Failed to apply preset");
      }

      // Close the form after successful application
      setTimeout(() => {
        process.exit(0);
      }, 1000);
    } catch (err) {
      await showToast({
        style: Toast.Style.Failure,
        title: "Preset Application Failed",
        message: err instanceof Error ? err.message : "Unknown error occurred",
      });
    }
  };

  const getPresetDescription = (presetName: string) => {
    switch (presetName) {
      case "minimal":
        return "Clean, minimal configuration with essential settings only";
      case "default":
        return "Balanced default configuration suitable for most users";
      case "developer":
        return "Developer-focused configuration with advanced features";
      default:
        return "Custom preset configuration";
    }
  };

  return (
    <Form
      actions={
        <ActionPanel>
          <Action
            title="Apply Preset"
            icon={Icon.CheckCircle}
            onAction={handleSubmit}
          />
        </ActionPanel>
      }
    >
      <Form.Description
        title={`Apply ${preset.charAt(0).toUpperCase() + preset.slice(1)} Preset`}
        text={`Application: ${app}\n\n${getPresetDescription(preset)}`}
      />
      <Form.Separator />
      <Form.Description text="This will apply the selected preset configuration to your application. Your current settings will be backed up automatically." />
    </Form>
  );
}
