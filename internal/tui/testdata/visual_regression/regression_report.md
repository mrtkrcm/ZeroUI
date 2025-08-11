# Visual Regression Test Report

Generated: 2025-08-12T05:26:54+07:00

## Test Summary

Total scenarios tested: 7

## Scenarios

### MainGrid_Standard
- **Description**: Main application grid in standard terminal size
- **Dimensions**: 120x40
- **Tolerance**: 1
- **Critical Elements**: [ZEROUI applications Ghostty VS Code]

### MainGrid_Small
- **Description**: Main application grid in small terminal
- **Dimensions**: 80x24
- **Tolerance**: 1
- **Critical Elements**: [ZEROUI applications]

### HelpOverlay_Standard
- **Description**: Help overlay display
- **Dimensions**: 120x40
- **Tolerance**: 0
- **Critical Elements**: [Help Navigation quit]

### ConfigEditor_View
- **Description**: Configuration editor interface
- **Dimensions**: 120x40
- **Tolerance**: 1
- **Critical Elements**: [Config ghostty]

### ErrorDisplay_Standard
- **Description**: Error message display
- **Dimensions**: 120x40
- **Tolerance**: 0
- **Critical Elements**: [Error test error message]

### ResponsiveLarge_160x50
- **Description**: Large terminal responsive layout
- **Dimensions**: 160x50
- **Tolerance**: 1
- **Critical Elements**: [ZEROUI applications 4 columns]

### ResponsiveNarrow_60x20
- **Description**: Very narrow terminal layout
- **Dimensions**: 60x20
- **Tolerance**: 2
- **Critical Elements**: [ZEROUI]

## Files Generated

- Current snapshots: `testdata/visual_regression/`
- Baseline images: `testdata/baseline_images/`
- Diff visualizations: `testdata/diff_images/`

