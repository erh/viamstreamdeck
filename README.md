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
```

### choose a font

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
   "brightness":100,
   "assets":{
      "fonts":[
         "/absolute/path/to/custom-font.ttf",
         "/Users/yourname/.fonts"
      ],
      "images":[
         "/absolute/path/to/custom-image.jpg",
         "/Users/yourname/images"
      ]
   },
   "keys":[
      {
         "key":0,
         "component":"my-component",
         "method":"do_command",
         "text":"Custom",
         "text_font":"custom-font.ttf",
         "image":"logo.png"
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

#### Addressing Keys by Name

A key can carry an optional `name` in its config. `update_display` accepts either the key's number or its name wherever it refers to a key, so callers don't depend on where a key sits on the deck.

```json
{
  "keys": [
    { "key": 0, "name": "status", "text": "Idle", "component": "foo", "method": "do_command" }
  ]
}
```

```json
{
  "update_display": {
    "keys": {
      "status": { "text": "Running", "color": "green" }
    }
  }
}
```

- With a `keys` list, a reference that starts with a number is that key number, as before, so `"4abc"` is key 4. Anything else is looked up as a name, so a name that starts with a number can't be used.
- With `pages`, only a whole number is a key number. Anything else is looked up as a name, so `"4abc"` is an error unless a key has that name, and a key named `"4"` can only be reached as key 4.
- Names are only looked up among the keys of the page or key list being updated, and a name used by more than one key there is an error.
- `name` is optional and is never validated when the config loads, so it cannot make an existing config fail. It cannot be changed by `update_display`.
- A reference that is neither a key number nor the name of a key is an error.

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

#### Dial example

```json
{
  "dials": [
    {
      "dial": 0,
      "component": "foo",
      "command": "SetPosition"
    }
  ]
}
```

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

### Pages

Instead of a flat list of keys, you can organize keys into named pages and switch between them at runtime using `set_page`. Use `initial_page` to set which page is displayed on startup. Keys can navigate between pages by referencing the Stream Deck component itself and calling `set_page`.

```json
{
  "brightness": 100,
  "initial_page": "main",
  "pages": {
    "main": [
      {
        "key": 0,
        "text": "Go to other",
        "color": "blue",
        "component": "my-streamdeck",
        "method": "do_command",
        "args": [{ "set_page": "other" }]
      },
      {
        "key": 1,
        "text": "Hello",
        "component": "foo",
        "method": "do_command",
        "args": [{ "x": 1 }]
      }
    ],
    "other": [
      {
        "key": 0,
        "text": "Back",
        "color": "red",
        "component": "my-streamdeck",
        "method": "do_command",
        "args": [{ "set_page": "main" }]
      },
      {
        "key": 1,
        "text": "Do thing",
        "component": "bar",
        "method": "do_command",
        "args": [{ "y": 2 }]
      }
    ]
  }
}
```

You cannot use `keys` and `pages` at the same time.

#### Updating pages with update_display

With `pages` configured, `update_display` edits a page instead of the flat key list, and `set_page` keeps working afterwards. Add `page` to choose which page to update. Without it, the update goes to the page that is showing.

```json
{
  "update_display": {
    "page": "other",
    "keys": {
      "1": { "text": "Done", "color": "green" }
    }
  }
}
```

- A change to the page that is showing is drawn immediately. A change to any other page is kept and drawn when that page is shown, so a page you can't see stays current.
- The update is checked when it is sent. If any key could not be drawn, for example an unknown `image`, an unsupported `method`, a `SetPosition` key whose position doesn't exist, or a key with nothing to display, the whole update is rejected before anything is drawn or changed, instead of failing later when someone switches to that page. If the deck itself fails partway through an update, for example because it is unplugged, the keys already applied stay applied. A key that points at a component that isn't present is not an error; it shows the crossed-circle image, as it does on every page.
- A key can only be added to the page that is showing. Updating a key that a hidden page doesn't have is an error.
- Updates last until the component is reconfigured, which resets the pages to the configured ones, as with flat `keys`.
- `brightness` and `dials` apply to the whole deck. `page` only affects `keys`, and is ignored by a config that uses `keys` instead of `pages`.
- The response is the same as for flat configs: the key numbers that were updated.

#### Behavior changes

A config that uses `keys` behaves as it did before. One thing is different:

| Case | Before | Now |
| --- | --- | --- |
| `update_display` on a config with `pages` | Added the key to a flat `keys` list, which switched the deck out of pages mode and made `set_page` fail | Edits the page and leaves pages mode alone |

Configs that load today still load unchanged. Every new field (`name`, and `page` on `update_display`) is optional.

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
