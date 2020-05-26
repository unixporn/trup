# Environment Variables

```sh
TOKEN=your_token
ROLE_MOD=707318869445967872
ROLE_COLORS=707318869445967872,707318869445967872
CHANNEL_SHOWCASE=635625917623828520
CHANNEL_FEEDBACK=
CHANNEL_BOTLOG=
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
- [davidv171](https://github.com/davidv171) & [GaugeK](https://github.com/GaugeK) for fetcher.sh
- [aosync](https://github.com/aosync) for commands purge and move
- [tteeoo](https://github.com/tteeoo) for commands git, desc and dotfiles
- [kayew](https://github.com/kayew) for a more explicit setfetch message
