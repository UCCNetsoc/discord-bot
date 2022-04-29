<img src="https://raw.githubusercontent.com/UCCNetsoc/wiki/master/assets/logo-service-discord-bot.svg" width="360"/>

[![CI](https://github.com/UCCNetsoc/discord-bot/actions/workflows/main.yml/badge.svg)](https://github.com/UCCNetsoc/discord-bot/actions/workflows/main.yml)

The UCC Netsoc discord bot has the following features:

- Provide events and announcements to the website.

- Provide daily updates on covid stats and vaccines.

- Provide upcoming events via a command.

- Shorten URLs using our internal URL shortener.

- Check if websites are up.

- And much more!

## Why make a new Discord Bot?

When we decided to implement the features allowing the posting of events/announcements, we realised that the bot could no longer be server agnostic.

For that reason we decided to reimplement the bot with the ability be configured with servers/channels which require elevated permissions.

## Running locally

1. Ensure you have both docker and docker-compose installed.

1. Ensure to clone this repo and the Netsoc [dev-env](https://github.com/UCCNetsoc/dev-env).

1. In the dev-env, run `./start-discord-bot.sh /path/to/this-repo` and follow the on screen prompts
