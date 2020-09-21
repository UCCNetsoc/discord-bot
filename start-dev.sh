#!/bin/bash

if [[ $# -eq 0 ]] ; then
    echo 'Must provide path to dev-env. For example: ./start-dev.sh ../dev-env'
    exit 1
fi

WD=`pwd`
cd $1

if [ ! -f "./discord-bot/docker-compose.override.yml" ]; then
    echo "Discord Bot Token from https://discord.com/developers/applications:" 
    read DISCORD
    echo "Sendgrid Token (optional, press enter if none)"
    read SENDGRID

	echo "version: \"3.7\" 
services:
  discord-bot:
    environment:
      - DISCORD_TOKEN=${DISCORD}
      - SENDGRID_TOKEN=${SENDGRID}
    volumes:
      - ${WD}:/bot
" > ./discord-bot/docker-compose.override.yml

    echo " Go to consul at http://localhost:8500 and set the following K/Vs in \`discordbot/\`:
   - \`channels\`: \`{\"public_announcements\": \"id\", \"private_events\": \"id\"}\`
   - \`servers\`: \`{\"public\": \"id\", \"committee\": \"id\"}\`
   - \`welcome_messages\`: \`[\"Hi %s, whats up\", \"Yo %s\"]\`
   "
fi

bash -c "./dev-env up discord-bot"
