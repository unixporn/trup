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

## Kudos to:
- [davidv171](https://github.com/davidv171) & [6gk](https://github.com/6gk) for fetcher.sh
- [aosync](https://github.com/aosync) for commands purge and move
- [tteeoo](https://github.com/tteeoo) for commands git, desc and dotfiles
- [kayew](https://github.com/kayew) for a more explicit setfetch message
- [davidv171](https://github.com/davidv171) for mute command
- [elkowar](https://github.com/elkowar) for blocklist command, media-logging in botlog and !poll multi
