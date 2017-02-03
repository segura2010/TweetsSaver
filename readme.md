# TweetsSaver

TweetsSaver is a little tool I wrote in Golang to download tweets using the Twitter REST API.

It downloads tweets which containt hastags or words configured in "query" field of the config.json file (see config_example.json). I had to anylize tweets in a specific location, so I wrote it to download tweets which are published in specific location (you can configure it on "location" field in config.json).

It downloads and save the tweets on a MongoDB database.

You can configure all you want by using the config.json file, and then run the program by running:

`bin/TweetsSaver -c path/to/config.json`

### Installation

1. Clone this repository and rename the folder to TweetsSaver if it is not the name.
2. Install and configure MongoDB.
3. Run the install script install.sh to install all the dependencies.
4. Compile using Makefile make.

Finally, run using the executable for your platform.

### TelegramBot

Telegram Bot is optional, you can keep these fields empty, but if you want to use a telegram bot to receive alerts when an error occurs, you must create a bot and put the token and your UserID in the config.json
