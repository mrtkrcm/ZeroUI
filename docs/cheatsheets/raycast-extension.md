# Raycast Extension Cheatsheet

**Documentation:** https://developers.raycast.com/
**Repository:** https://github.com/raycast/extensions
**Purpose:** Build native macOS extensions with React, TypeScript, and Node.js

## Quick Reference

### Development Commands
```bash
npm run dev          # Start development mode with hot reload
npm run build        # Production build with type checking
npm run lint         # Run ESLint checks
npm run fix-lint     # Auto-fix lint issues
npm run publish      # Publish to store
```

### Essential Imports
```typescript
// Core API
import {
  List, Detail, Form, Action, ActionPanel,
  showToast, Toast, getPreferenceValues,
  environment, LaunchType
} from "@raycast/api";

// Utilities (separate package)
import {
  usePromise, useCachedPromise, useCachedState,
  useExec, useFetch, useForm
} from "@raycast/utils";
```

## Core Components

### List
The primary UI component for displaying collections of items.
```typescript
export default function Command() {
  const [searchText, setSearchText] = useState("");

  return (
    <List
      isLoading={isLoading}
      searchText={searchText}
      onSearchTextChange={setSearchText}
      searchBarPlaceholder="Search items..."
      filtering={true}  // Built-in filtering by item titles
    >
      <List.Section title="Section Name">
        <List.Item
          title="Item Title"
          subtitle="Optional subtitle"
          icon={Icon.Star}
          accessories={[
            { text: "metadata" },
            { icon: Icon.Circle, tooltip: "Status" }
          ]}
          actions={
            <ActionPanel>
              <Action title="Do Something" onAction={() => {}} />
            </ActionPanel>
          }
        />
      </List.Section>
    </List>
  );
}
```

### Detail
Render markdown content with optional metadata panel.
```typescript
<Detail
  markdown={markdownContent}
  navigationTitle="Page Title"
  metadata={
    <Detail.Metadata>
      <Detail.Metadata.Label title="Status" text="Active" />
      <Detail.Metadata.Link
        title="URL"
        target="https://example.com"
        text="Open Link"
      />
      <Detail.Metadata.Separator />
      <Detail.Metadata.TagList title="Tags">
        <Detail.Metadata.TagList.Item text="Tag1" color={Color.Blue} />
      </Detail.Metadata.TagList>
    </Detail.Metadata>
  }
/>
```

### Form
Collect user input with validation.
```typescript
export default function Command() {
  return (
    <Form
      actions={
        <ActionPanel>
          <Action.SubmitForm title="Submit" onSubmit={handleSubmit} />
        </ActionPanel>
      }
    >
      <Form.TextField
        id="name"
        title="Name"
        placeholder="Enter name"
        error={errors.name}
        onChange={() => setErrors({ ...errors, name: undefined })}
      />
      <Form.Dropdown id="type" title="Type" defaultValue="option1">
        <Form.Dropdown.Item value="option1" title="Option 1" />
        <Form.Dropdown.Item value="option2" title="Option 2" />
      </Form.Dropdown>
      <Form.Checkbox id="enabled" label="Enable feature" defaultValue={true} />
      <Form.TextArea id="description" title="Description" />
      <Form.DatePicker id="date" title="Date" type={Form.DatePicker.Type.Date} />
      <Form.FilePicker id="file" title="File" allowMultipleSelection={false} />
    </Form>
  );
}
```

### Actions & ActionPanel
Define user interactions.
```typescript
<ActionPanel>
  <ActionPanel.Section title="Primary">
    <Action
      title="Custom Action"
      icon={Icon.Star}
      shortcut={{ modifiers: ["cmd"], key: "s" }}
      onAction={() => console.log("clicked")}
    />
    <Action.OpenInBrowser url="https://example.com" />
    <Action.CopyToClipboard content="text to copy" />
    <Action.Paste content="text to paste" />
    <Action.Push title="Show Detail" target={<DetailView />} />
  </ActionPanel.Section>
  <ActionPanel.Section title="Secondary">
    <Action.ShowInFinder path="/path/to/file" />
    <Action.Open target="/path/to/file" />
    <Action.Trash paths={["/path/to/file"]} />
  </ActionPanel.Section>
</ActionPanel>
```

## React Hooks (@raycast/utils)

### usePromise
Handle async functions with loading/error states.
```typescript
const { data, isLoading, error, revalidate } = usePromise(
  async (query: string) => {
    const response = await fetch(`/api/search?q=${query}`);
    return response.json();
  },
  [searchText],  // Dependencies
  {
    execute: searchText.length > 0,  // Conditional execution
    onData: (data) => console.log("Loaded:", data),
    onError: (error) => showToast({ style: Toast.Style.Failure, title: "Error" }),
  }
);
```

### useCachedPromise
Stale-while-revalidate caching strategy.
```typescript
const { data, isLoading, revalidate } = useCachedPromise(
  async (app: string) => {
    return await fetchAppConfig(app);
  },
  [appName],
  {
    keepPreviousData: true,  // Prevent flicker on param change
    initialData: [],         // Default value
  }
);
```

### useCachedState
Persist state between command runs.
```typescript
// Persisted across sessions
const [favorites, setFavorites] = useCachedState<string[]>("favorites", []);

// Shared across commands using same key
const [theme, setTheme] = useCachedState<string>("app-theme", "dark");
```

### useExec
Execute shell commands.
```typescript
const { data, isLoading, error } = useExec("zeroui", ["list", "apps"], {
  parseOutput: ({ stdout }) => stdout.trim().split("\n"),
  onError: (error) => {
    showToast({ style: Toast.Style.Failure, title: "Command failed" });
  },
});
```

### useFetch
Fetch remote APIs.
```typescript
const { data, isLoading } = useFetch<ApiResponse>(
  `https://api.example.com/data?q=${query}`,
  {
    keepPreviousData: true,
    headers: { Authorization: `Bearer ${token}` },
  }
);
```

### useForm
Form validation and state management.
```typescript
const { handleSubmit, itemProps, values, setValidationError } = useForm<FormValues>({
  onSubmit: async (values) => {
    await saveData(values);
    showToast({ title: "Saved!" });
  },
  validation: {
    name: (value) => {
      if (!value?.length) return "Name is required";
      if (value.length < 3) return "Minimum 3 characters";
    },
    email: (value) => {
      if (!value?.includes("@")) return "Invalid email";
    },
  },
});

return (
  <Form actions={...}>
    <Form.TextField title="Name" {...itemProps.name} />
    <Form.TextField title="Email" {...itemProps.email} />
  </Form>
);
```

## Manifest (package.json)

### Command Configuration
```json
{
  "commands": [
    {
      "name": "list-items",
      "title": "List Items",
      "subtitle": "My Extension",
      "description": "Shows all items",
      "mode": "view",
      "keywords": ["search", "items"],
      "icon": "list-icon.png"
    },
    {
      "name": "quick-action",
      "title": "Quick Action",
      "mode": "no-view",
      "description": "Performs action without UI"
    },
    {
      "name": "menubar-status",
      "title": "Status",
      "mode": "menu-bar",
      "interval": "10m",
      "description": "Shows status in menu bar"
    }
  ]
}
```

### Mode Types
| Mode | Description |
|------|-------------|
| `view` | Shows UI (List, Detail, Form) |
| `no-view` | Runs without UI (API calls, clipboard) |
| `menu-bar` | Menu bar item with dropdown |

### Background Refresh Intervals
```
"interval": "10s"   // 10 seconds (minimum)
"interval": "5m"    // 5 minutes
"interval": "1h"    // 1 hour
"interval": "1d"    // 1 day
```

### Preferences
```json
{
  "preferences": [
    {
      "name": "apiKey",
      "title": "API Key",
      "description": "Your API key",
      "type": "password",
      "required": true
    },
    {
      "name": "theme",
      "title": "Theme",
      "type": "dropdown",
      "required": false,
      "default": "dark",
      "data": [
        { "title": "Light", "value": "light" },
        { "title": "Dark", "value": "dark" }
      ]
    },
    {
      "name": "enableCache",
      "title": "Enable Cache",
      "type": "checkbox",
      "default": true,
      "label": "Cache API responses"
    },
    {
      "name": "configPath",
      "title": "Config File",
      "type": "file",
      "required": false
    }
  ]
}
```

### Preference Types
| Type | Description |
|------|-------------|
| `textfield` | Single-line text input |
| `password` | Secure text input |
| `checkbox` | Boolean toggle |
| `dropdown` | Select from options |
| `appPicker` | Select installed app |
| `file` | File picker |
| `directory` | Directory picker |

### Command Arguments
```json
{
  "commands": [{
    "name": "search",
    "arguments": [
      {
        "name": "query",
        "type": "text",
        "placeholder": "Search query",
        "required": true
      },
      {
        "name": "limit",
        "type": "text",
        "placeholder": "Max results",
        "required": false
      }
    ]
  }]
}
```

## Environment & Lifecycle

### Environment API
```typescript
import { environment } from "@raycast/api";

// Extension info
environment.extensionName     // "my-extension"
environment.commandName       // "list-items"
environment.commandMode       // "view" | "no-view" | "menu-bar"

// Launch context
environment.launchType        // LaunchType.UserInitiated | LaunchType.Background
environment.launchContext     // Custom data from launchCommand()

// Paths
environment.assetsPath        // Path to assets/ directory
environment.supportPath       // Path for extension data storage

// Raycast info
environment.raycastVersion    // "1.60.0"
environment.isDevelopment     // true in dev mode
environment.theme             // "light" | "dark"
```

### Launch Types
```typescript
import { environment, LaunchType } from "@raycast/api";

export default function Command() {
  if (environment.launchType === LaunchType.Background) {
    // Triggered by background refresh or launchCommand
    return performBackgroundTask();
  }

  // Normal user-initiated launch
  return <List>...</List>;
}
```

### Cross-Command Communication
```typescript
import { launchCommand, LaunchType } from "@raycast/api";

// Launch another command with context
await launchCommand({
  name: "detail-view",
  type: LaunchType.UserInitiated,
  context: { itemId: "123", refresh: true },
});

// In target command, access via props
export default function Command(props: LaunchProps) {
  const { itemId, refresh } = props.launchContext || {};
}
```

### Deeplinks
```
raycast://extensions/{author}/{extension}/{command}?context={json}
raycast://extensions/mrtkrcm/zeroui/list-apps?context={"app":"ghostty"}
```

## Error Handling

### Toast Notifications
```typescript
import { showToast, Toast } from "@raycast/api";

// Success
await showToast({
  style: Toast.Style.Success,
  title: "Saved",
  message: "Configuration updated",
});

// Error with action
await showToast({
  style: Toast.Style.Failure,
  title: "Failed to load",
  message: error.message,
  primaryAction: {
    title: "Retry",
    onAction: () => revalidate(),
  },
});

// Animated loading
const toast = await showToast({
  style: Toast.Style.Animated,
  title: "Processing...",
});

// Update toast when done
toast.style = Toast.Style.Success;
toast.title = "Done!";
```

### Graceful Degradation
```typescript
const { data, error, isLoading } = useCachedPromise(fetchData, [], {
  onError: (error) => {
    // Don't disrupt user flow for non-critical errors
    console.error("Fetch failed:", error);
    showToast({ style: Toast.Style.Failure, title: "Using cached data" });
  },
});

// Show cached/stale data if available, even on error
return <List isLoading={isLoading}>
  {(data || cachedData).map(item => <List.Item key={item.id} {...item} />)}
</List>;
```

### Alert Dialogs
```typescript
import { confirmAlert, Alert } from "@raycast/api";

const confirmed = await confirmAlert({
  title: "Delete Item",
  message: "This action cannot be undone.",
  icon: Icon.Trash,
  primaryAction: {
    title: "Delete",
    style: Alert.ActionStyle.Destructive,
  },
});

if (confirmed) {
  await deleteItem();
}
```

## Debugging

### Console Logging
```typescript
// Visible in terminal during `npm run dev`
console.log("Debug:", variable);
console.warn("Warning:", message);
console.error("Error:", error);

// Note: console.log is disabled in store builds
```

### VSCode Debugger
1. Install "Raycast" VSCode extension
2. Add breakpoints in code
3. Run `npm run dev`
4. Attach debugger from VSCode

### React DevTools
```bash
# Terminal 1: Start DevTools
npx react-devtools

# Terminal 2: Run extension
npm run dev
# DevTools connects automatically
```

### Debug Log Collection
```bash
# Stream Raycast debug logs
log stream --predicate "subsystem == 'com.raycast.macos'" --level debug --style compact >> ~/Desktop/raycast-debug.log
```

### Extension Issues Dashboard
View crash reports and errors for published extensions:
https://www.raycast.com/extension-issues

### Development vs Production
```typescript
if (environment.isDevelopment) {
  console.log("Debug info:", data);
}
```

## Store Publishing Checklist

### Requirements
- [ ] Custom 512x512px PNG icon (not default Raycast icon)
- [ ] `package-lock.json` included (use npm, not yarn/pnpm)
- [ ] `npm run build` succeeds locally
- [ ] `npm run lint` passes
- [ ] README.md with setup instructions (if API keys needed)
- [ ] Screenshots (recommended)

### package.json Required Fields
```json
{
  "name": "my-extension",
  "title": "My Extension",
  "description": "Clear description of what it does",
  "icon": "extension-icon.png",
  "author": "your-username",
  "license": "MIT",
  "categories": ["Productivity", "Developer Tools"],
  "commands": [...]
}
```

### Review Process
- Initial review within 5 business days
- PRs created as draft for you to add description
- Significant changes require original author sign-off
- Follow community guidelines and acceptable use policy

## Recent API Changes (2024-2025)

### AI Tools (v1.89+)
New entry point type for AI-powered extensions:
```json
{
  "tools": [{
    "name": "search-docs",
    "title": "Search Documentation",
    "description": "Search the docs for information"
  }]
}
```

### useLocalStorage Hook
Simplified local storage management:
```typescript
import { useLocalStorage } from "@raycast/utils";

const { value, setValue, removeValue, isLoading } = useLocalStorage<string[]>(
  "recent-items",
  []
);
```

### AI Model Enum
```typescript
import { AI } from "@raycast/api";

await AI.ask("Summarize this", { model: AI.Model.Anthropic_Claude_Sonnet });
// Available: OpenAI_GPT4o, Anthropic_Claude_Sonnet, etc.
```

### Pagination Support
All data hooks now support pagination:
```typescript
const { data, pagination } = useCachedPromise(
  async (cursor) => fetchPage(cursor),
  [],
  {
    // Return { data: items[], hasMore: boolean }
  }
);

return (
  <List pagination={pagination}>
    {data?.map(item => <List.Item key={item.id} />)}
  </List>
);
```

### Draft PR Workflow
When publishing, PRs are created as drafts so you can fill in the description before submitting for review.

## Useful Resources

- [API Documentation](https://developers.raycast.com/)
- [Extension Store](https://raycast.com/store)
- [GitHub Discussions](https://github.com/raycast/extensions/discussions)
- [Extension Guidelines](https://manual.raycast.com/extensions)
- [Changelog](https://developers.raycast.com/misc/changelog)
