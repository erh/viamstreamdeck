# Module viam-streamdeck 

Integration with Elgato StreamDeck

## simple-config

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
                 "args": {
                     "x ": 1
                 }
             },
             ...
             }
}
```

## pickup

This is a simple streamdeck app for picking things up

```
{
  "brightness" : 100, // optional, 0 -> 100
  "arm" : "...",
  "gripper" : "...",
  "finder" : "...", // vision service that supports GetObjectPointClouds
  "motion" : "...",  // motion service, could be 'builtin'
  "watch_pose" : "...", // an arm-saver position from where to watch
}
```
