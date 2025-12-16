import asyncio
import logging
import os
from datetime import datetime
from json import dumps
from telethon import __version__
from telethon.sessions import MemorySession
from telethon.sync import TelegramClient
from telethon.tl.alltlobjects import LAYER

try:
    import uvloop
    uvloop.install()
except ImportError:
    pass

logging.basicConfig(
    level=logging.INFO,
    format="[%(asctime)s - %(levelname)s] - %(name)s - %(message)s",
    datefmt="%d-%b-%y %H:%M:%S",
    handlers=[
        logging.StreamHandler()
    ]
)


APP_ID = int(os.environ.get("APP_ID", "6"))
API_HASH = os.environ.get("API_HASH", "")
BOT_TOKEN = os.environ.get("BOT_TOKEN")
FLOOD_SLEEP_THRESHOLD = int(os.environ.get("FLOOD_WAIT_SLEEP_TIME", "10"))
MESSAGE_LINK = os.environ.get("MESSAGE_LINK", "")


async def main():
    d = {}
    async with TelegramClient(
        session=MemorySession(),
        api_id=APP_ID,
        api_hash=API_HASH,
        flood_sleep_threshold=FLOOD_SLEEP_THRESHOLD,
        receive_updates=False
    ) as app:
        d["version"] = __version__
        d["layer"] = LAYER
        await app.start(bot_token=BOT_TOKEN)

        # Flexible parsing for t.me/c/ID/MSG_ID (private) vs t.me/User/MSG_ID (public)
        parts = MESSAGE_LINK.strip().rstrip("/").split("/")
        if len(parts) >= 3 and parts[-3] == "c":
            chat_id = int("-100" + parts[-2])
            s_message_id = int(parts[-1])
        else:
            chat_id = parts[-2]
            s_message_id = int(parts[-1])

        t1 = datetime.now()
        message = await app.get_messages(entity=chat_id, ids=int(s_message_id))
        d["file_size"] = message.file.size
        t2 = datetime.now()
        filename = await message.download_media()
        t3 = datetime.now()
        d["download"] = {
            "start_time": t2.timestamp(),
            "end_time": t3.timestamp(),
            "time_taken": (t3 - t2).seconds
        }
        t4 = datetime.now()
        await app.send_file(
            entity=int(chat_id),
            file=filename,
            caption="Telethon",
            force_document=True,
            reply_to=message
        )
        t5 = datetime.now()
        d["upload"] = {
            "start_time": t4.timestamp(),
            "end_time": t5.timestamp(),
            "time_taken": (t5 - t4).seconds
        }
        os.remove(filename)

    print(dumps(d, indent=2))


if __name__ == "__main__":
    asyncio.run(main())
