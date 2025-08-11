# Component Consolidation Plan

## Current Duplicate Components

1. **Grid Components**
   - `AppGridModel` (legacy) - /components/app_grid.go
   - `HuhGridModel` (modern) - /components/huh_grid.go
   - Decision: Keep HuhGridModel, migrate AppGridModel features

2. **Config Editor Components**  
   - `ConfigEditorModel` (legacy) - /components/config_editor.go
   - `HuhConfigEditorModel` (modern) - /components/huh_config_editor.go
   - Decision: Keep HuhConfigEditorModel with form validation

3. **App Selector Components**
   - `AppSelectorModel` (legacy) - /components/app_selector.go
   - `HuhAppSelectorModel` (modern) - /components/huh_app_selector.go
   - Decision: Keep HuhAppSelectorModel with better filtering

## Migration Steps

1. Update app.go to use only Huh components
2. Rename Huh components to remove "Huh" prefix
3. Delete legacy components
4. Update all imports and references
5. Test functionality

## Benefits
- 50% less code to maintain
- Consistent architecture
- Better form handling with Huh
- Unified styling approach