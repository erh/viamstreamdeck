from typing import ClassVar, Final, Mapping, Optional, Sequence, Tuple
from google.protobuf import json_format

from StreamDeck.DeviceManager import DeviceManager

from typing_extensions import Self
from viam.proto.app.robot import ComponentConfig
from viam.proto.common import ResourceName
from viam.resource.base import ResourceBase
from viam.resource.easy_resource import EasyResource
from viam.resource.types import Model, ModelFamily
from viam.services.generic import *
from viam.utils import ValueTypes

def key_change_callback(deck, key, state):
    print("Deck {} Key {} = {}".format(deck.id(), key, state), flush=True)

class StreamdeckOriginal(Generic, EasyResource):
    # To enable debug-level logging, either run viam-server with the --debug option,
    # or configure your resource/machine to display debug logs.
    MODEL: ClassVar[Model] = Model(
        ModelFamily("erh", "viam-streamdeck"), "streamdeck-original"
    )

    @classmethod
    def new(cls, config: ComponentConfig, dependencies: Mapping[ResourceName, ResourceBase]) -> Self:
        service = super().new(config, dependencies)

        streamdecks = DeviceManager().enumerate()

        self.logging.info("Found {} Stream Deck(s).".format(len(streamdecks)))

        for index, deck in enumerate(streamdecks):
            # This example only works with devices that have screens.
            if not deck.is_visual():
                continue

            deck.open()
            deck.reset()

            self.logger.info("Opened '{}' device (serial number: '{}', fw: '{}')".format(
                deck.deck_type(), deck.get_serial_number(), deck.get_firmware_version()
            ))
            
            deck.set_key_callback(key_change_callback)
            self.deck = deck
            break

        self.reconfigure(config, dependencies)
        return service

    

    @classmethod
    def validate_config(cls, config: ComponentConfig) -> Tuple[Sequence[str], Sequence[str]]:
        return [], []

    def reconfigure(self, config: ComponentConfig, dependencies: Mapping[ResourceName, ResourceBase]):

        deck.set_brightness(30)

        
        python_dict = json_format.MessageToDict(config.attributes)
        self.logger.info(python_dict)

        return super().reconfigure(config, dependencies)

    async def do_command(self,command: Mapping[str, ValueTypes],*,timeout: Optional[float] = None,**kwargs) -> Mapping[str, ValueTypes]:
        self.logger.error("`do_command` is not implemented")
        raise NotImplementedError()

