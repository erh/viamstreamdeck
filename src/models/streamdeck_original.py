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

#temp hack
from viam.components.arm import ArmClient
from viam.components.base import BaseClient
from viam.components.button import ButtonClient
from viam.components.camera import CameraClient
from viam.components.gripper import GripperClient
from viam.components.motor import MotorClient
from viam.components.sensor import SensorClient
from viam.components.switch import SwitchClient


logger = getLogger(__name__)

def get_attributes(config: ComponentConfig):
    return struct_to_dict(config.attributes)

def get_keys(attrs):
    if "keys" in attrs:
        return attrs["keys"]
    return []

class StreamdeckOriginal(Generic, EasyResource):
    # To enable debug-level logging, either run viam-server with the --debug option,
    # or configure your resource/machine to display debug logs.
    MODEL: ClassVar[Model] = Model(
        ModelFamily("erh", "viam-streamdeck"), "streamdeck-original"
    )

    def __init__(self, x):
        super().__init__(x)
        self.deck = None
        self.dependencies = None
    
    @classmethod
    def new(cls, config: ComponentConfig, dependencies: Mapping[ResourceName, ResourceBase]) -> Self:
        return super().new(config, dependencies)

    @classmethod
    def validate_config(cls, config: ComponentConfig) -> Tuple[Sequence[str], Sequence[str]]:
        attrs = get_attributes(config)
        return cls.validate_attrs(attrs)

    @classmethod
    def validate_attrs(cls, attrs) -> Tuple[Sequence[str], Sequence[str]]:
        keys = get_keys(attrs)

        componentNames = []
        for k in keys:
            for req in ["component", "key", "text", "method", "args"]:
                if req not in k:
                    raise ValidationError("need a component for all keys, missing %s for %s" % (req, str(k)))
            
            cn = k["component"]
            if cn not in componentNames:
                componentNames.append(cn)

            knumber = int(k["key"])

        return componentNames, []
        
    
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

        self.dependencies = dependencies
                
        keys = get_keys(attrs)
        for k in keys:
            self.logger.info("key {}".format(k))
            color = "black"
            if "color" in k:
                color = k["color"]
            image = PILHelper.create_key_image(self.deck, background=color)
            draw = ImageDraw.Draw(image)
            draw.text((image.width / 2, image.height / 2), text=k["text"], anchor="ms", fill="white", font_size=15)
            kf = PILHelper.to_native_key_format(self.deck, image)
            self.deck.set_key_image(int(k["key"]), kf)
        self.keys = keys

    async def key_change_callback(self, deck, key, state):
        if not state:
            return
        for k in self.keys:
            if key == int(k["key"]):
                await self.key_press(k)
                return
        self.logger.info("no mapping for key: {}".format(key))


    async def key_press(self, key_info):
        self.logger.info("key press {}".format(key_info))
        cn = key_info["component"]
        if self.dependencies is None:
            self.logger.info("no dependencies at all, testing?")
            return
        for d in self.dependencies:
            if str(d).endswith("/" + cn): #TODO is this correct???
                return self.key_press_component(key_info, d, self.dependencies[d])
        self.logger.error("could not find dependency for %s" % cn)
        

    async def key_press_component(self, key_info, theName, theResource):
        self.logger.info("key press component {} {} {}".format(key_info, theName, theResource))
        m = theResource.__getattribute__(key_info["method"])
        result = await m(*key_info["args"])
        self.logger.info("result {}".format(result))

        
    async def do_command(self,command: Mapping[str, ValueTypes],*,timeout: Optional[float] = None,**kwargs) -> Mapping[str, ValueTypes]:
        logger.error("`do_command` is not implemented")
        raise NotImplementedError()


    async def close(self):
        if self.deck is not None:
            self.deck.close()
            self.deck = None
        await super().close()


class SillyForTest:
    def __init__(self, x):
        self.x = x
        
    async def do_command(self,command):
        return {"x" : self.x, "cmd" : command}
        
async def quick_test():
    c = {"brightness" : 100,
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
             {
                 "text": "cat",
                 "key": 1,
                 "color" : "purple",
                 "component": "foo",
                 "method": "do_command",
                 "args": {
                     "x ": 1
                 }
             },
             {
                 "text": "is",
                 "key": 2,
                 "color" : "purple",
                 "component": "foo",
                 "method": "do_command",
                 "args": {
                     "x ": 1
                 }
             },
             {
                 "text": "in",
                 "key": 3,
                 "color" : "purple",
                 "component": "foo",
                 "method": "do_command",
                 "args": {
                     "x ": 1
                 }
             },
             {
                 "text": "the",
                 "key": 4,
                 "color" : "purple",
                 "component": "foo",
                 "method": "do_command",
                 "args": {
                     "x ": 1
                 }
             },
             {
                 "text": "box.",
                 "key": 5,
                 "color" : "purple",
                 "component": "foo",
                 "method": "do_command",
                 "args": {
                     "x ": 1
                 }
             },
             {
                 "text": "cat",
                 "key": 6,
                 "color" : "purple",
                 "component": "foo",
                 "method": "do_command",
                 "args": {
                     "x ": 1
                 }
             },
             {
                 "text": "cats",
                 "key": 7,
                 "color" : "purple",
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
                 "color" : "blue",
                 "method": "do_command",
                 "args": {
                     "x ": 1
                 }
             },
             {
                 "text": "eliot",
                 "key": 11,
                 "component": "bar",
                 "color" : "green",
                 "method": "do_command",
                 "args": {
                     "x ": 1
                 }
             }

         ],
         }
    sd = StreamdeckOriginal("x")
    print(StreamdeckOriginal.validate_attrs(c))
    sd.reconfigure2(None, c, {"asd/bar" : SillyForTest(5), "asd/foo" : SillyForTest(7)})
    await asyncio.sleep(5)
    await sd.close()

if __name__ == '__main__':
    asyncio.run(quick_test())

