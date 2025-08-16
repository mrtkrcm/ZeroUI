#!/usr/bin/env python3
import json
import re
import yaml
import urllib.request
from typing import Any, Dict, List

def fetch_default_json():
    """Fetch Zed's default settings JSON from GitHub"""
    url = "https://raw.githubusercontent.com/zed-industries/zed/main/assets/settings/default.json"
    with urllib.request.urlopen(url) as response:
        content = response.read().decode('utf-8')
    
    # Remove comments (// style)
    lines = content.split('\n')
    cleaned_lines = []
    for line in lines:
        # Remove // comments but preserve URLs like "https://"
        if '//' in line:
            # Check if it's a URL
            if 'http://' in line or 'https://' in line:
                # Keep the line as is if it contains a URL
                cleaned_lines.append(line)
            else:
                # Remove comment part
                comment_pos = line.find('//')
                cleaned_lines.append(line[:comment_pos])
        else:
            cleaned_lines.append(line)
    
    cleaned_content = '\n'.join(cleaned_lines)
    
    # Remove trailing commas before closing brackets/braces
    cleaned_content = re.sub(r',(\s*[}\]])', r'\1', cleaned_content)
    
    # Parse JSON
    return json.loads(cleaned_content)

def infer_type(value: Any) -> str:
    """Infer the type of a setting value"""
    if isinstance(value, bool):
        return "boolean"
    elif isinstance(value, (int, float)):
        return "number"
    elif isinstance(value, str):
        return "string"
    elif isinstance(value, list):
        return "array"
    elif isinstance(value, dict):
        return "object"
    else:
        return "string"

def flatten_settings(obj: Dict, prefix: str = "") -> Dict[str, Dict]:
    """Flatten nested settings into a flat dictionary"""
    settings = {}
    
    for key, value in obj.items():
        full_key = f"{prefix}.{key}" if prefix else key
        
        # Skip complex nested objects for now, but include their basic form
        if isinstance(value, dict) and not any(k in ['mode', 'light', 'dark', 'provider', 'model'] for k in value.keys()):
            # Recursively flatten nested objects
            nested = flatten_settings(value, full_key)
            settings.update(nested)
        else:
            # Add the setting
            setting_type = infer_type(value)
            description = f"Configuration for {full_key.replace('_', ' ').replace('.', ' ')}"
            
            # Determine category
            category = "general"
            if "font" in full_key or "text" in full_key:
                category = "font"
            elif "theme" in full_key or "color" in full_key or "appearance" in full_key:
                category = "appearance"
            elif "panel" in full_key or "dock" in full_key or "tab" in full_key:
                category = "ui"
            elif "git" in full_key:
                category = "git"
            elif "agent" in full_key or "ai" in full_key:
                category = "ai"
            elif "language" in full_key or "lsp" in full_key:
                category = "language"
            elif "key" in full_key or "vim" in full_key or "helix" in full_key:
                category = "keybindings"
            elif "scroll" in full_key or "cursor" in full_key:
                category = "editor"
            elif "diagnostic" in full_key:
                category = "diagnostics"
            elif "format" in full_key or "indent" in full_key:
                category = "formatting"
            
            settings[full_key] = {
                "name": full_key,
                "type": setting_type,
                "description": description,
                "category": category
            }
            
            # Add default value if it's not null
            if value is not None:
                if setting_type == "string" and value:
                    settings[full_key]["default_value"] = value
                elif setting_type in ["number", "boolean"]:
                    settings[full_key]["default_value"] = value
                elif setting_type == "array" and value:
                    settings[full_key]["default_value"] = value
    
    return settings

def generate_reference_yaml():
    """Generate a complete Zed reference configuration YAML"""
    print("Fetching Zed default settings...")
    default_settings = fetch_default_json()
    
    print("Flattening settings...")
    flattened = flatten_settings(default_settings)
    
    # Create the reference config structure
    reference = {
        "app_name": "Zed",
        "config_path": "~/.config/zed/settings.json",
        "config_type": "json",
        "settings": flattened
    }
    
    # Write to YAML
    output_path = "/Users/m/code/muka-hq/configtoggle/configs/zed.yaml"
    with open(output_path, 'w') as f:
        yaml.dump(reference, f, default_flow_style=False, sort_keys=False, allow_unicode=True)
    
    print(f"Generated {len(flattened)} settings in {output_path}")
    
    # Print summary
    categories = {}
    for setting in flattened.values():
        cat = setting.get('category', 'general')
        categories[cat] = categories.get(cat, 0) + 1
    
    print("\nSettings by category:")
    for cat, count in sorted(categories.items()):
        print(f"  {cat}: {count}")

if __name__ == "__main__":
    generate_reference_yaml()