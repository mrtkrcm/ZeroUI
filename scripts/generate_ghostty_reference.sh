# DRY_RUN HANDLER
if [ "${DRY_RUN:-0}" != "0" ]; then
  echo "(DRY-RUN) $0: DRY_RUN enabled, skipping destructive actions."
fi

#!/bin/bash

# Generate complete Ghostty reference config from the actual Ghostty documentation

echo "app_name: \"Ghostty\""
echo "config_path: \"~/.config/ghostty/config\""
echo "config_type: \"custom\""
echo ""
echo "settings:"

# Extract all config options with their documentation
ghostty +show-config --default --docs 2>/dev/null | awk '
BEGIN {
    desc = ""
    multiline_desc = ""
    option_name = ""
    default_value = ""
    in_option = 0
}

# Capture description lines (lines starting with #)
/^# / {
    line = substr($0, 3)
    if (line != "") {
        if (desc == "") {
            desc = line
        } else {
            multiline_desc = multiline_desc " " line
        }
    }
    next
}

# Capture config option lines
/^[a-z-]+ = / {
    # If we have a previous option, print it
    if (option_name != "") {
        print_option()
    }
    
    # Parse new option
    option_name = $1
    default_value = substr($0, length(option_name) + 4)
    in_option = 1
    next
}

# Handle empty lines - reset description capture
/^$/ {
    if (in_option && option_name != "") {
        print_option()
    }
    desc = ""
    multiline_desc = ""
    option_name = ""
    default_value = ""
    in_option = 0
    next
}

function print_option() {
    # Skip if option name is empty
    if (option_name == "") return
    
    # Handle duplicate keys (like palette) by adding a suffix
    if (seen[option_name]++) {
        option_name = option_name "-" seen[option_name]
    }
    
    # Clean up option name for YAML
    yaml_name = option_name
    gsub("-", "_", yaml_name)
    
    print "  " option_name ":"
    print "    name: \"" option_name "\""
    
    # Determine type based on default value and description
    type = "string"
    if (default_value == "true" || default_value == "false") {
        type = "boolean"
    } else if (default_value ~ /^[0-9]+$/ || default_value ~ /^[0-9]+\.[0-9]+$/) {
        type = "number"
    } else if (desc ~ /color/ || option_name ~ /color/ || option_name ~ /background/ || option_name ~ /foreground/) {
        type = "color"
    } else if (desc ~ /choose|select|one of/ || desc ~ /Valid values/) {
        type = "choice"
    }
    
    print "    type: \"" type "\""
    
    # Add description
    full_desc = desc
    if (multiline_desc != "") {
        full_desc = full_desc " " multiline_desc
    }
    # Escape quotes in description
    gsub(/"/, "\\\"", full_desc)
    # Limit description length
    if (length(full_desc) > 200) {
        full_desc = substr(full_desc, 1, 197) "..."
    }
    if (full_desc != "") {
        print "    description: \"" full_desc "\""
    }
    
    # Add default value if present
    if (default_value != "") {
        if (type == "boolean") {
            print "    default_value: " default_value
        } else if (type == "number") {
            print "    default_value: " default_value
        } else if (default_value != "\"\"" && default_value != "") {
            gsub(/"/, "\\\"", default_value)
            print "    default_value: \"" default_value "\""
        }
    }
    
    # Determine category based on option name
    category = "general"
    if (option_name ~ /^font-/ || option_name ~ /adjust-.*thickness/) {
        category = "font"
    } else if (option_name ~ /color|background|foreground|palette|theme/) {
        category = "appearance"
    } else if (option_name ~ /^window-/ || option_name ~ /fullscreen|maximize/) {
        category = "window"
    } else if (option_name ~ /^cursor-/) {
        category = "cursor"
    } else if (option_name ~ /^mouse-/) {
        category = "mouse"
    } else if (option_name ~ /scroll/) {
        category = "scrolling"
    } else if (option_name ~ /^clipboard-|^copy-/) {
        category = "clipboard"
    } else if (option_name ~ /^keybind/) {
        category = "keybindings"
    } else if (option_name ~ /^macos-/) {
        category = "macos"
    } else if (option_name ~ /^gtk-|^linux-/) {
        category = "linux"
    } else if (option_name ~ /shell|command|working-directory/) {
        category = "shell"
    }
    
    print "    category: \"" category "\""
    print ""
}

END {
    # Print last option if exists
    if (option_name != "") {
        print_option()
    }
}
'
