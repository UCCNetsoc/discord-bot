# Netsoc Discord Bot
[![](https://ci.netsoc.co/api/badges/uccnetsoc/discord-bot/status.svg)](https://ci.netsoc.co/UCCNetsoc/discord-bot/)

The UCC Netsoc discord bot has the following features:
  - Allows UCC students to automatically register as a member for the public Discord Server
  
  - Allows committee members to publish events and announcements to multiple mediums using one command. This includes:
    - The public discord server
    - The website (coming soon)
    
  - Recall events/announcements from these platforms after being sent
  
## Why make a new Discord Bot
When we decided to implement the features allowing the posting of events/announcements, we realised that the bot could no longer be server agnostic.

For that reason we decided to reimplement the bot with the ability to watch/read our consul K/V store to allow for real time configuration of what servers/channels require elevated permissions.

## Running locally
1. Ensure you have both docker and docker-compose installed.

2. Ensure to clone this repo and the Netsoc [dev-env](https://github.com/UCCNetsoc/dev-env).

3. cd into `dev-env/discord-bot` and make a file called `docker-compose.override.yml`. Ensure it contains the following:
  ```yml
  version: "3.7"
services:
  discord-bot:
    environment:
      - DISCORD_TOKEN=Put discord token here
    volumes:
      - /path/to/discord-bot/repo:/bot
  ```
  
4. cd into the `dev-env` and run `./dev-env up discord-bot consul`.
