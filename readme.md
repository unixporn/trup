# Requirements

- Go
- PostgreSQL 9.1 or up

# Setup

```sh
# Database
createdb trup
psql trup <db/structure.sql
```

# Environment Variables

```sh
DATABASE_URL=postgresql://user@localhost/trup
TOKEN=your_token
CHANNEL_SHOWCASE=635625917623828520
ROLE_MOD=707318869445967872 # moderator role for modping
```

## Kudos to:
- davidv7 & @GaugeK for fetcher.sh