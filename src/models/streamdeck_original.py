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
        return [], []


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
        self.reconfigure2(self, config, attrs, dependencies)
        
    def reconfigure2(self, config: ComponentConfig, attrs, dependencies: Mapping[ResourceName, ResourceBase]):
        self.find_deck()

        b = float(attrs.get("brightness", 50))
        logger.info("setting brightness to: {}".format(b))
        self.deck.set_brightness(b)

        keys = []
        if "keys" in attrs:
            keys = attrs["keys"]
        for k in keys:
            image = PILHelper.create_key_image(self.deck)
            draw = ImageDraw.Draw(image)
            draw.text((image.width / 2, image.height / 2), text=k["text"], anchor="ms", fill="white")
            kf = PILHelper.to_native_key_format(self.deck, image)
            self.deck.set_key_image(k["key"], kf)
        self.keys = keys

    def key_change_callback(self, deck, key, state):
        for k in self.keys:
            if key == k["key"]:
                print("Key {} = {}".format(k, state), flush=True)


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
    sd.reconfigure2(None,
                    {"brightness" : 100,
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
                     },
                    None)
    await asyncio.sleep(5)
    await sd.close()

if __name__ == '__main__':
    asyncio.run(quick_test())

