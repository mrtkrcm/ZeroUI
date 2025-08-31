/// <reference types="@raycast/api">

/* 🚧 🚧 🚧
 * This file is auto-generated from the extension's manifest.
 * Do not modify manually. Instead, update the `package.json` file.
 * 🚧 🚧 🚧 */

/* eslint-disable @typescript-eslint/ban-types */

type ExtensionPreferences = {
  /** ZeroUI Binary Path - Path to the ZeroUI binary. Leave empty to use default location. */
  "zerouiPath"?: string,
  /** Command Timeout - Timeout for ZeroUI commands in milliseconds. */
  "timeout": string,
  /** Enable Caching - Cache results to improve performance. Disable if you need real-time data. */
  "enableCache": boolean,
  /** Cache Duration - How long to cache results in milliseconds. */
  "cacheDuration": string
}

/** Preferences accessible in all the extension's commands */
declare type Preferences = ExtensionPreferences

declare namespace Preferences {
  /** Preferences accessible in the `list-apps` command */
  export type ListApps = ExtensionPreferences & {}
  /** Preferences accessible in the `toggle-config` command */
  export type ToggleConfig = ExtensionPreferences & {}
  /** Preferences accessible in the `list-values` command */
  export type ListValues = ExtensionPreferences & {}
  /** Preferences accessible in the `list-changed` command */
  export type ListChanged = ExtensionPreferences & {}
  /** Preferences accessible in the `keymap-list` command */
  export type KeymapList = ExtensionPreferences & {}
  /** Preferences accessible in the `manage-presets` command */
  export type ManagePresets = ExtensionPreferences & {}
  /** Preferences accessible in the `cache-management` command */
  export type CacheManagement = ExtensionPreferences & {}
  /** Preferences accessible in the `menubar` command */
  export type Menubar = ExtensionPreferences & {}
}

declare namespace Arguments {
  /** Arguments passed to the `list-apps` command */
  export type ListApps = {}
  /** Arguments passed to the `toggle-config` command */
  export type ToggleConfig = {}
  /** Arguments passed to the `list-values` command */
  export type ListValues = {}
  /** Arguments passed to the `list-changed` command */
  export type ListChanged = {}
  /** Arguments passed to the `keymap-list` command */
  export type KeymapList = {}
  /** Arguments passed to the `manage-presets` command */
  export type ManagePresets = {}
  /** Arguments passed to the `cache-management` command */
  export type CacheManagement = {}
  /** Arguments passed to the `menubar` command */
  export type Menubar = {}
}

