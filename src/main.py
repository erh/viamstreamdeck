import asyncio
from viam.module.module import Module
try:
    from models.streamdeck_original import StreamdeckOriginal
except ModuleNotFoundError:
    # when running as local module with run.sh
    from .models.streamdeck_original import StreamdeckOriginal


if __name__ == '__main__':
    asyncio.run(Module.run_from_registry())
