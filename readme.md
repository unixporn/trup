# Environment Variables

```sh
TOKEN=your_token
ROLE_MOD=707318869445967872
ROLE_MUTE=
ROLE_COLORS=707318869445967872,707318869445967872
CHANNEL_SHOWCASE=
CATEGORY_MOD_PRIVATE=
CHANNEL_FEEDBACK=
CHANNEL_MODLOG=
CHANNEL_AUTO_MOD=
CHANNEL_BOT_MESSAGES=
CHANNEL_BOT_TRAFFIC=
```

# Setup with Docker

After cloning the repository, create a file called `.env` 
containing the necessary environment variables (as shown above) in the project root.

Afterwards you can initialize the docker services by running
```sh
docker-compose up -d
```

# Automatic setup with Nix

```sh
# Install the Nix package manager:
curl -L https://nixos.org/nix/install | sh
# Clone this repo
git clone https://github.com/unixporn/trup
cd trup
# Enter the dev environment
nix-shell
# remember to set your Environment Variables
go run .
# All done, the bot should be running now.
```

# Requirements

- Go
- PostgreSQL 11 or up

# Setup the Database

No need to do this if you use Nix.

```sh
# Database
createdb trup
psql trup <db/structure.sql
export DATABASE_URL=postgresql://user@localhost/trup
```

# Bot Commands 

Bot flag is !. 
The list of bot commands is : 
- dotfiles - Adds a dotfiles link to your fetch.
- git - Adds a git link to your fetch.
- setfetch - Run wihout arguments to see instructions.
- role - Use without arguments to see available roles.
- poll - Creates a poll.
- repo - Sends a link to the bot's repository.
- info - Displays additional informations.
- pfp Displays the user's profile picture in highest resolution.
- modping - Pings online mods. (Don't abuse.)
- desc - Sets or clears your descriptions, displays it with fetch.

## Kudos to:
- [davidv171](https://github.com/davidv171) & [GaugeK](https://github.com/GaugeK) for fetcher.sh
- [aosync](https://github.com/aosync) for commands purge and move
- [tteeoo](https://github.com/tteeoo) for commands git, desc and dotfiles
- [kayew](https://github.com/kayew) for a more explicit setfetch message
- [davidv171](https://github.com/davidv171) for mute command
- [elkowar](https://github.com/elkowar) for blocklist command, media-logging in botlog and !poll multi
