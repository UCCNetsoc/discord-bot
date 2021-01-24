<img src="https://raw.githubusercontent.com/UCCNetsoc/wiki/master/assets/logo-service-discord-bot.svg" width="360"/>


[![](https://ci.netsoc.co/api/badges/UCCNetsoc/discord-bot/status.svg)](https://ci.netsoc.co/UCCNetsoc/discord-bot)

The UCC Netsoc discord bot has the following features:

- Allows UCC students to automatically register as a member for the public Discord Server

- Allows committee members to publish events and announcements to multiple mediums using one command. This includes:
  - The public discord server
  - The website
  - Twitter
- Recall events/announcements from these platforms after being sent

- And much more!

## Why make a new Discord Bot?

When we decided to implement the features allowing the posting of events/announcements, we realised that the bot could no longer be server agnostic.

For that reason we decided to reimplement the bot with the ability be configured with servers/channels which require elevated permissions.

## Running locally

1. Ensure you have both docker and docker-compose installed.

1. Ensure to clone this repo and the Netsoc [dev-env](https://github.com/UCCNetsoc/dev-env).

1. In this repo, run `./start-dev.sh /path/to/dev-env` and follow the on screen prompts
