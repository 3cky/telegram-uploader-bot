Assumptions of the service file
-------------------------------

1. `telegram-uploader-bot` binary lives in `/usr/local/bin` directory.
2. Bot is run as root by default (can be changed in service file).
3. Bot configuration file is in `/usr/local/etc/telegram-uploader-bot.cfg`.
4. Logs are written to stdout/journald.

How to use these files
----------------------

Place service file to systemd directory: `sudo cp telegram-uploader-bot.service /etc/systemd/system/`

Set service file ownership to root: `sudo chown root:root /etc/systemd/system/telegram-uploader-bot.service`

Reload systemd daemon: `sudo systemctl daemon-reload`

Start telegram-uploader-bot service: `sudo systemctl start telegram-uploader-bot`

Reload telegram-uploader-bot config: `sudo systemctl reload telegram-uploader-bot`

Enable telegram-uploader-bot service to run on boot: `sudo systemctl enable telegram-uploader-bot`
