# Trup

**A Discord bot for the Unixporn community**

![build status](https://github.com/unixporn/trup/workflows/build/badge.svg)
![lint status](https://github.com/unixporn/trup/workflows/lint/badge.svg)

# Dependencies

- Go
- PostgreSQL 11 or above

# Setup with Docker

After cloning the repository, create a file called `.env` containing the necessary environment variables (as shown in `/.env.example`) in the project root.

Afterwards, you can initialize the Docker services by running

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

# Set environment variables
cp .env.example .env

go run .
```

# Set up the database

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
- [Tanish2002](https://github.com/Tanish2002) for info command, cleaner !help, DM for ban and blocklist violation and more options for role command
