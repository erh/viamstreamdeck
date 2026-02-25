# Module viam-streamdeck 

Integration with Elgato StreamDeck

## attributes

### simple-config

```json
{
  "brightness" : 100, // 0 -> 100
  "keys": [
             {
                 "text": "The",
                 "key": 0,
                 "color" : "purple",
                 "component": "foo",
                 "method": "do_command",
                 "args": [ {
                     "x ": 1
                 } ]
             }
  ]
}

### choosing a font
```json
{
  "brightness" : 100, // 0 -> 100
  "keys": [
             {
                 "text": "The",
                 "text_font": "NotoEmoji-Regular.tff"
             }
  ]
}
```

Currently only `NotoEmoji-Regular.tff` is included with the module. You can load additional fonts as an asset if desired. Note that rendering is limited to a small character set within hte font. This is a known limitation of the freetype library's DrawString() method that is used. Cahracters outside the Basic Multilingual Plane (BMP) (characters above U+FFFF) cannot be rendered. If this is needed, use images instead.


### adding images and font assets

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

You can add your own external fonts and images to use througout your configuration. Fonts must be in `.ttf` or `.otf`. Images can be `.jpg`, `.jpeg`, `.png` or `.gif`. Images `stopsign.jpg` and `x.jpg ` is included and can also be used without an external asset.

### DoCommand: update_display

The `update_display` DoCommand allows you to dynamically update the Stream Deck display at runtime. This is useful for changing key appearances, updating brightness, or modifying dial configurations without restarting the component.

#### Updating Brightness

```json
{
  "update_display": {
    "brightness": 75
  }
}
```

Changes the Stream Deck brightness to 75% (accepts values 0-100).

#### Updating Keys

You can update individual keys by specifying the key number and the properties to change. The update merges with the existing configuration, so you only need to specify the properties you want to change.

```json
{
  "update_display": {
    "keys": {
      "0": {
        "text": "New Text",
        "color": "blue"
      },
      "1": {
        "text": "Status: OK",
        "text_color": "green",
        "image": "checkmark.png"
      }
    }
  }
}
```

**Available key properties:**
- `text` - Text to display on the key
- `text_color` - Color of the text (e.g., "red", "blue", "#FF0000")
- `color` - Background color of the key
- `image` - Image file to display (must be in assets)
- `component` - Component to call when key is pressed
- `method` - Method to call on the component
- `args` - Array of arguments to pass to the method

#### Updating Dials

For Stream Decks with dials (like the Stream Deck+), you can update dial configurations:

```json
{
  "update_display": {
    "dials": {
      "0": {
        "component": "my-motor",
        "command": "set_power"
      }
    }
  }
}
```

**Available dial properties:**
- `component` - Component to call when dial is turned
- `command` - Command to execute on the component

#### Combining Updates

You can update multiple aspects in a single DoCommand:

```json
{
  "update_display": {
    "brightness": 80,
    "keys": {
      "0": {
        "text": "Active",
        "color": "green"
      },
      "2": {
        "text": "Inactive",
        "color": "gray"
      }
    },
    "dials": {
      "0": {
        "component": "volume-control",
        "command": "set_level"
      }
    }
  }
}
```

#### Response

The DoCommand returns a map indicating what was updated:

```json
{
  "brightness": 80,
  "keys": [0, 2],
  "dials": [0]
}
```

#### Self-Referencing Keys

Keys can reference the Stream Deck component itself by using the component's own name. This allows keys to trigger DoCommands that update the display dynamically:

```json
{
  "keys": [
    {
      "key": 0,
      "text": "Toggle",
      "component": "my-streamdeck",
      "method": "do_command",
      "args": [
        {
          "update_display": {
            "keys": {
              "1": {
                "text": "Updated!",
                "color": "yellow"
              }
            }
          }
        }
      ]
    }
  ]
}
```

## pickup

This is a simple streamdeck app for picking things up

```json
{
  "brightness" : 100, // optional, 0 -> 100
  "arm" : "...",
  "gripper" : "...",
  "finder" : "...", // vision service that supports GetObjectPointClouds
  "motion" : "...",  // motion service, could be 'builtin'
  "watch_pose" : "...", // an arm-saver position from where to watch
}
```
