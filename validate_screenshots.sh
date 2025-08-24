#!/bin/bash

echo "🔍 ZeroUI Screenshot System Validation"
echo "====================================="

# Check if testdata directory exists
if [ ! -d "testdata" ]; then
    echo "❌ testdata directory not found"
    exit 1
fi

# Check for screenshots directory
if [ ! -d "testdata/screenshots" ]; then
    echo "📁 Creating screenshots directory..."
    mkdir -p testdata/screenshots
fi

echo "📁 Screenshots directory structure:"
find testdata/screenshots -type d 2>/dev/null | sort

echo ""
echo "📄 Screenshot files found:"
find testdata/screenshots -type f \( -name "*.html" -o -name "*.json" -o -name "*.txt" \) 2>/dev/null | sort

echo ""
echo "📊 Screenshot count by type:"
echo "HTML files: $(find testdata/screenshots -name "*.html" 2>/dev/null | wc -l)"
echo "JSON files: $(find testdata/screenshots -name "*.json" 2>/dev/null | wc -l)"
echo "TXT files:  $(find testdata/screenshots -name "*.txt" 2>/dev/null | wc -l)"

echo ""
echo "📋 Demo files:"
if [ -d "testdata/screenshots/demo" ]; then
    echo "✅ Demo directory exists"
    ls -la testdata/screenshots/demo/
else
    echo "❌ Demo directory not found"
fi

echo ""
echo "📋 Manual demo files:"
if [ -d "testdata/screenshots/manual_demo" ]; then
    echo "✅ Manual demo directory exists"
    ls -la testdata/screenshots/manual_demo/
else
    echo "❌ Manual demo directory not found"
fi

echo ""
echo "🎯 Screenshot System Status:"
if [ -f "internal/tui/screenshot_demo_test.go" ]; then
    echo "✅ Standalone screenshot test exists"
else
    echo "❌ Standalone screenshot test missing"
fi

if [ -f "screenshot_demo.go" ]; then
    echo "✅ Manual screenshot demo exists"
else
    echo "❌ Manual screenshot demo missing"
fi

if [ -f "testdata/README_SCREENSHOTS.md" ]; then
    echo "✅ Screenshot documentation exists"
else
    echo "❌ Screenshot documentation missing"
fi

echo ""
echo "🚀 Next Steps:"
echo "1. Run: go run screenshot_demo.go"
echo "2. Check: testdata/screenshots/manual_demo/"
echo "3. View: Open HTML files in browser"
echo "4. Test: go test ./internal/tui -run TestScreenshotDemo -v"

echo ""
echo "✅ Validation complete!"
