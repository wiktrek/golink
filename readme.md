# Golink

Simple fast link shortener.

## Setup

Use SQLite for a quick local setup:

```sh
sqlite3 golink.db < schema.sqlite.sql
DATABASE_URL='sqlite://golink.db?_pragma=foreign_keys(1)'
```

Or use MySQL:

```sh
mysql -u USER -p DATABASE < schema.sql
DATABASE_URL='mysql://USER:PASSWORD@127.0.0.1:3306/DATABASE?parseTime=true'
```
Put `DATABASE_URL` into a `.env` file

The app listens on `:8080` by default (`ADDR` changes this). The dashboard is
at `/dashboard`; links use `/go/{slug}`.

## Separate link domain

To create links such as `example.link/my-link` while keeping the dashboard on
`example.com`, set:

```sh
LINK_DOMAIN=example.link
```

Point both domains at the app. To use a separate redirect-only port instead,
also set `LINK_ADDR` (for example, `:8081`).

# TODO
- Change password
- track how long a request took
- track how long routes take to respond
<!-- color palette: https://www.realtimecolors.com/?colors=eaf5f2-0a1814-93dac5-257e65-51d4af&fonts=Inter-Inter -->