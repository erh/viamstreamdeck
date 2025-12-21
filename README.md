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
