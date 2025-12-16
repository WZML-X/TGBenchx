import asyncio
import logging
import os
from datetime import datetime
from json import dumps
from pyrogram import Client, __version__
from pyrogram.raw.all import layer


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
SLEEP_THRESHOLD = int(os.environ.get("FLOOD_WAIT_SLEEP_TIME", "10"))
TG_SESSION = os.environ.get("TG_SESSION", "")
MESSAGE_LINK = os.environ.get("MESSAGE_LINK", "")


async def main():
    d = {}
    async with Client(
        name="my_account",
        session_string=TG_SESSION,
        in_memory=True,
        api_id=APP_ID,
        api_hash=API_HASH,
        sleep_threshold=SLEEP_THRESHOLD,
        no_updates=True,
        bot_token=BOT_TOKEN
    ) as app:
        app.upload_boost = True

        d["version"] = __version__
        d["layer"] = layer

        # Flexible parsing for t.me/c/ID/MSG_ID (private) vs t.me/User/MSG_ID (public)
        parts = MESSAGE_LINK.strip().rstrip("/").split("/")
        if len(parts) >= 3 and parts[-3] == "c":
            # Private chat/channel: t.me/c/12345/67 -> chat_id="-10012345"
            chat_id = int("-100" + parts[-2])
            s_message_id = int(parts[-1])
        else:
            # Public chat: t.me/username/67 -> chat_id="username"
            chat_id = parts[-2]
            s_message_id = int(parts[-1])

        t1 = datetime.now()
        message = await app.get_messages(chat_id=chat_id, message_ids=int(s_message_id))
        d["file_size"] = message.document.file_size
        t2 = datetime.now()
        filename = await message.download()
        t3 = datetime.now()
        d["download"] = {
            "start_time": t2.timestamp(),
            "end_time": t3.timestamp(),
            "time_taken": (t3 - t2).seconds
        }
        t4 = datetime.now()
        await app.send_document(
            chat_id=message.chat.id,
            document=filename,
            caption="Pyrogram",
            reply_to_message_id=message.id
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
    try:
        import uvloop
        uvloop.install()
    except ImportError:
        pass
    asyncio.run(main())
