# telegram-uploader-bot

Telegram bot written in Go that monitors directories for new files and uploads them to chats.

## Features

- Monitors specified directories for new files and uploads them to Telegram chats.
- Supports multiple directories and chats configuration.
- Can add custom tags to uploaded files using plain text tags, regexps, or [expr](https://github.com/antonmedv/expr) language.
- File filtering using file masks.

## Prerequisites

- Go (v1.20 or higher) installed on your Linux system.
- A Telegram bot token obtained by creating a new bot using [BotFather](https://core.telegram.org/bots#how-do-i-create-a-bot).

## Installation

```bash
git clone https://github.com/3cky/telegram-uploader-bot.git
cd telegram-uploader-bot
make install
```

## Configuration

The bot uses a YAML config file set by `-c` or `--config-file` flags. The default config file is "/usr/local/etc/telegram-uploader-bot.cfg".

To use a custom configuration file, run the bot with the `-c` flag followed by the path to your config file:

```bash
telegram-uploader-bot -c /path/to/custom_config.cfg
```

Example configuration:

```yaml
telegram:
  token: "my-telegram-bot-token"

uploads:
  - directory: "/path/to/watch/dir"
    files:
      - "*.jpg" # case insensitive match
    documents: false # set to true to upload files as documents (without reencoding)
    chat: 1234567
    tags:
      plain:
        - "work"
        - "important"
      regexp:
        - ".*/(?P<name>.*)\\.jpg" # matched groups will be used as tags prefixed by group names
      expr:
        - "(file.Size() > 1024 * 1024) ? 'big' : ''" # tag files bigger than 1 megabyte
```

## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

## License

telegram-uploader-bot is released under the Apache 2.0 license. See the [LICENSE](https://github.com/3cky/telegram-uploader-bot/blob/main/LICENSE) file for more info.
