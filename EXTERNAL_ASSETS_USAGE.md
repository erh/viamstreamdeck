# External Assets Configuration

This guide explains how to load custom fonts and images from disk.

## External Assets (Fonts and Images)

You can load fonts and images from absolute file paths on disk, in addition to the embedded assets.

### Configuration

Add an `assets` section to your Viam config with lists of absolute paths (files or folders):

```json
{
  "services": [
    {
      "name": "streamdeck",
      "api": "rdk:service:generic",
      "model": "michaellee1019:viam-streamdeck:streamdeck-original",
      "attributes": {
        "brightness": 100,
        "assets": {
          "fonts": [
            "/absolute/path/to/custom-font.ttf",
            "/Users/yourname/.fonts"
          ],
          "images": [
            "/absolute/path/to/custom-image.jpg",
            "/Users/yourname/images"
          ]
        },
        "keys": [
          {
            "key": 0,
            "component": "my-component",
            "method": "do_command",
            "text": "Custom",
            "text_font": "custom-font.ttf",
            "image": "logo.png"
          }
        ]
      }
    }
  ]
}
```

### How It Works

1. **External assets are loaded at initialization** - The module loads fonts and images from the specified paths
2. **Supports files and folders** - Each path can be either:
   - A specific file: `/path/to/font.ttf`
   - A folder: `/path/to/fonts/` (loads all supported files in the folder)
3. **Filtered by extension** - When loading from folders:
   - Fonts: Only `.ttf` and `.otf` files are loaded
   - Images: Only `.jpg`, `.jpeg`, `.png`, and `.gif` files are loaded
4. **Combined with embedded assets** - External assets are added to the built-in embedded assets
5. **Referenced by filename** - Use the base filename (not full path) in `text_font` and `image` fields
6. **Reloaded on reconfigure** - When you update the config, external assets are reloaded

### Example: Adding Custom Fonts

```json
{
  "assets": {
    "fonts": [
      "/home/user/.fonts/NotoEmoji-Regular.ttf",
      "/home/user/custom-fonts/MyFont-Bold.ttf"
    ]
  },
  "keys": [
    {
      "key": 0,
      "text": "✅",
      "text_font": "NotoEmoji-Regular.ttf",
      "component": "sensor",
      "method": "do_command"
    }
  ]
}
```

### Example: Adding Custom Images

```json
{
  "assets": {
    "images": [
      "/var/lib/images/warning-icon.png",
      "/home/user/logos/company.jpg"
    ]
  },
  "keys": [
    {
      "key": 1,
      "image": "warning-icon.png",
      "text": "Alert",
      "component": "sensor",
      "method": "do_command"
    }
  ]
}
```

### Example: Loading from Folders

```json
{
  "assets": {
    "fonts": [
      "/home/user/.fonts",
      "/usr/share/fonts/custom"
    ],
    "images": [
      "/home/user/streamdeck-icons",
      "/var/lib/custom-images"
    ]
  },
  "keys": [
    {
      "key": 0,
      "text": "✅",
      "text_font": "NotoEmoji-Regular.ttf",
      "component": "sensor",
      "method": "do_command"
    },
    {
      "key": 1,
      "image": "icon-warning.png",
      "component": "sensor",
      "method": "do_command"
    }
  ]
}
```

This will load:
- All `.ttf` and `.otf` files from `/home/user/.fonts` and `/usr/share/fonts/custom`
- All `.jpg`, `.jpeg`, `.png`, and `.gif` files from the image directories
- Reference them by filename in your key configs

## Complete Example

Here's a complete config with custom assets:

```json
{
  "components": [
    {
      "name": "my-sensor",
      "api": "rdk:component:sensor",
      "model": "some:model:sensor"
    }
  ],
  "services": [
    {
      "name": "streamdeck",
      "api": "rdk:service:generic",
      "model": "michaellee1019:viam-streamdeck:streamdeck-original",
      "attributes": {
        "brightness": 100,
        "assets": {
          "fonts": [
            "/home/user/.fonts"
          ],
          "images": [
            "/home/user/icons"
          ]
        },
        "keys": [
          {
            "key": 0,
            "text": "✅",
            "text_font": "NotoEmoji-Regular.ttf",
            "color": "green",
            "component": "my-sensor",
            "method": "do_command",
            "args": [{"action": "start"}]
          },
          {
            "key": 1,
            "text": "❌",
            "text_font": "NotoEmoji-Regular.ttf",
            "color": "red",
            "component": "my-sensor",
            "method": "do_command",
            "args": [{"action": "stop"}]
          },
          {
            "key": 2,
            "image": "warning.png",
            "text": "Alert",
            "text_color": "yellow",
            "component": "my-sensor",
            "method": "do_command",
            "args": [{"action": "alert"}]
          }
        ]
      }
    }
  ]
}
```

## Notes

### Supported Font Formats
- `.ttf` - TrueType Font
- `.otf` - OpenType Font

### Supported Image Formats
- `.jpg` / `.jpeg`
- `.png`
- `.gif`

### File Paths
- Must be **absolute paths** (starting with `/` on Linux/Mac, or drive letter on Windows)
- Can be **individual files** or **directories**
- When using directories:
  - Only files with supported extensions are loaded
  - Subdirectories are NOT scanned (only files in the specified directory)
  - Fonts: `.ttf`, `.otf`
  - Images: `.jpg`, `.jpeg`, `.png`, `.gif`
- Files must be readable by the module process
- Filenames must be unique (the base filename is used as the key)

### Text Rendering
- Default font size is 20pt (automatically adjusts if text is too long)
- Uses the built-in MonoRegular font by default
- Specify `text_font` to use a custom font

### Emoji Support Limitations
Remember that freetype only supports emojis in the U+2xxx range:
- ✅ ❌ ⚠️ ⭐ ✨ ⚡ ❤️ ▶️ ⏸️ ☑️ ➕ ➖ (these work!)
- 😀 😬 👍 🚗 (these show as boxes - outside BMP)

See the main documentation for more details on emoji limitations.

### Troubleshooting

**Error: "failed to load external assets"**
- Check that all file paths exist and are absolute
- Verify the module has read permissions for the files
- Check file formats are supported

**Font/Image not found**
- Make sure you're using the **filename only** in `text_font` and `image`, not the full path
- Example: Use `"MyFont.ttf"` not `"/path/to/MyFont.ttf"`

**Emoji displays as box**
- The emoji is outside the Basic Multilingual Plane (above U+FFFF)
- Use emojis in the U+2xxx range instead
- See "Emoji Support Limitations" above for supported emojis
