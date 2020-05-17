# Environment Variables

```sh
TOKEN=your_token
CHANNEL_SHOWCASE=635625917623828520
ROLE_MOD=707318869445967872
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
- davidv7 & @GaugeK for fetcher.sh
