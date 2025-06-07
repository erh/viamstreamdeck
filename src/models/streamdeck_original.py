import asyncio
from typing import ClassVar, Final, Mapping, Optional, Sequence, Tuple
from google.protobuf import json_format

from PIL import Image, ImageDraw, ImageFont
from StreamDeck.DeviceManager import DeviceManager
from StreamDeck.ImageHelpers import PILHelper


from typing_extensions import Self
from viam.proto.app.robot import ComponentConfig
from viam.proto.common import ResourceName
from viam.resource.base import ResourceBase
from viam.resource.easy_resource import EasyResource
from viam.resource.types import Model, ModelFamily
from viam.services.generic import *
from viam.utils import ValueTypes, struct_to_dict
from viam.logging import getLogger

logger = getLogger(__name__)

def get_attributes(config: ComponentConfig):
    return struct_to_dict(config.attributes)

def get_keys(attrs):
    if "keys" in attrs:
        return attrs["keys"]
    return []

def extract_components(keys) -> Sequence[str]:
    d = {}
    for k in keys:
        d[k["component"]] = True

    a = []
    for k in d:
        a.append(k)
    return a

class StreamdeckOriginal(Generic, EasyResource):
    # To enable debug-level logging, either run viam-server with the --debug option,
    # or configure your resource/machine to display debug logs.
    MODEL: ClassVar[Model] = Model(
        ModelFamily("erh", "viam-streamdeck"), "streamdeck-original"
    )

    def __init__(self, x):
        super().__init__(x)
        self.deck = None
    
    @classmethod
    def new(cls, config: ComponentConfig, dependencies: Mapping[ResourceName, ResourceBase]) -> Self:
        return super().new(config, dependencies)

    @classmethod
    def validate_config(cls, config: ComponentConfig) -> Tuple[Sequence[str], Sequence[str]]:
        attsrs = get_attributes(config)
        keys = get_keys(attsrs)
        cs = extract_components(keys)
        logger.info("components: {}".format(cs))
        return cs, []
        
    def find_deck(self):
        if self.deck:
            return

        streamdecks = DeviceManager().enumerate()

        logger.info("Found {} Stream Deck(s).".format(len(streamdecks)))

        for index, deck in enumerate(streamdecks):
            # This example only works with devices that have screens.
            if not deck.is_visual():
                continue

            deck.open()
            deck.reset()

            logger.info("Opened '{}' device (serial number: '{}', fw: '{}')".format(
                deck.deck_type(), deck.get_serial_number(), deck.get_firmware_version()
            ))

            deck.set_key_callback(self.key_change_callback)
            self.deck = deck

            return

        raise "Did not find streamdecks"


    def reconfigure(self, config: ComponentConfig, dependencies: Mapping[ResourceName, ResourceBase]):
        attrs = get_attributes(config)
        logger.info("attributes: {}".format(attrs))
        self.reconfigure2(config, attrs, dependencies)
        
    def reconfigure2(self, config: ComponentConfig, attrs, dependencies: Mapping[ResourceName, ResourceBase]):
        self.find_deck()

        b = float(attrs.get("brightness", 50))
        logger.info("setting brightness to: {}".format(b))
        self.deck.set_brightness(b)

        if dependencies:
            print(dependencies)
            for x in dependencies:
                print(x)
                print(dependencies[x])
                
        keys = get_keys(attrs)
        for k in keys:
            self.logger.info("key {}".format(k))
            image = PILHelper.create_key_image(self.deck)
            draw = ImageDraw.Draw(image)
            draw.text((image.width / 2, image.height / 2), text=k["text"], anchor="ms", fill="white")
            kf = PILHelper.to_native_key_format(self.deck, image)
            self.deck.set_key_image(int(k["key"]), kf)
        self.keys = keys

    def key_press(self, key_info):
        self.logger.info("key pres {}".format(key_info))
        
    def key_change_callback(self, deck, key, state):
        if not state:
            return
        for k in self.keys:
            if key == k["key"]:
                self.key_press(k)
                return
        self.logger.info("no mapping for key: {}".format(key))


    async def do_command(self,command: Mapping[str, ValueTypes],*,timeout: Optional[float] = None,**kwargs) -> Mapping[str, ValueTypes]:
        logger.error("`do_command` is not implemented")
        raise NotImplementedError()


    async def close(self):
        if self.deck is not None:
            self.deck.close()
            self.deck = None
        await super().close()



async def quick_test():
    sd = StreamdeckOriginal("x")
    c = {"brightness" : 100,
         "keys": [
             {
                 "text": "foo",
                 "key": 0,
                 "component": "foo",
                 "method": "do_command",
                 "args": {
                     "x ": 1
                 }
             },
             {
                 "text": "bar",
                 "key": 8,
                 "component": "bar",
                 "method": "do_command",
                 "args": {
                     "x ": 1
                 }
             }
         ],
         }
    print(extract_components(c["keys"]))
    sd.reconfigure2(None, c, None)
    await asyncio.sleep(5)
    await sd.close()

if __name__ == '__main__':
    asyncio.run(quick_test())

