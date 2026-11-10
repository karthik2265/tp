# fh — feed handler

The feed handler reads historical NYSE trades and publishes them to the kdb+
ticker plant (tp), one trade at a time. It is the entry point of the
ticker plant setup: `fh → tp → rdb / rts`.

```
trades_data/20261012.json.zst ─┐
trades_data/20261013.json.zst ─┼─► replay ──(kdb+ IPC over TCP)──► tp  .u.upd[`trade; ...]
...                            ─┘
```

## Layout

```
fh/
├── cmd/replay/main.go   replay command: reads the files, sends trades to the tp
├── trades_data/         Databento trade files, one per day (YYYYMMDD.json.zst)
├── config/              settings from env vars / .env (LoadTP for replay)
├── .env.example         template for .env
├── go.mod / go.sum
└── README.md
```

Each folder under `cmd/` is its own program (`package main`). Run them from
`fh/` by folder path, e.g. `go run ./cmd/replay`.

## Data

`trades_data/` holds a Databento batch download:

| | |
|---|---|
| Dataset | `XNYS.PILLAR` (NYSE trades only, not the consolidated tape) |
| Schema | `trades` |
| Symbols | TSLA, SPY, MSFT, GOOGL |
| Dates | 2026-10-12 to 2026-11-09, one file per trading day (21 files, ~1.08M trades) |
| Details | `metadata.json`, `manifest.json`, `condition.json` |

Each `YYYYMMDD.json.zst` file is **newline-delimited JSON compressed with
zstd**: one JSON object per line, one line per trade.

```sh
zstd -dc trades_data/20261012.json.zst | head -1
```

```json
{"ts_recv":"2026-10-12T11:00:54.557393456Z","hd":{"ts_event":"2026-10-12T11:00:54.557182568Z","rtype":0,"publisher_id":9,"instrument_id":15144},"action":"T","side":"N","depth":0,"price":"778.780000000","size":38,"flags":128,"ts_in_delta":200242,"sequence":15918,"symbol":"SPY"}
```

The download used `pretty_px`, `pretty_ts` and `map_symbols`, so prices are
decimal **strings**, timestamps are ISO-8601 text, and every record carries its
`symbol`. Times are UTC and include pre-market trading.

## Running

Replay has no command-line flags. The tp address comes from environment
variables:

| Variable | Required | Example | Meaning |
|---|---|---|---|
| `TP_HOST` | yes | `localhost` | tp host |
| `TP_PORT` | yes | `5020` | tp port |
| `TP_AUTH` | no | `user:password` | only if the tp runs with `-u`/`-U` |

Set them in `fh/.env` (copy `.env.example`) or in the shell; shell values win
over `.env`. Without `TP_HOST`/`TP_PORT`, replay exits with
`Missing ticker plant address: set TP_HOST and TP_PORT`.

Start the tp first, then from `fh/`:

```sh
go run ./cmd/replay
TP_HOST=otherhost TP_PORT=5020 go run ./cmd/replay   # override for one run
```

Everything else is fixed in `cmd/replay/main.go`:

| Setting | Value |
|---|---|
| Data folder | `trades_data` (relative to `fh/`, so run from there) |
| tp table | `trade` |
| Rate | one trade per second |
| Days | all files in `trades_data`, earliest first |

Stop with Ctrl-C; replay exits cleanly after the current trade. At one trade
per second a full day takes hours and all 21 days take about 12 days.

## How the data is sent

### 1. Read

Files are replayed in date order, starting from the earliest. Each file is
streamed: zstd decompresses it chunk by chunk, each line is parsed with
`encoding/json`, and only the fields below are kept. Nothing is loaded into
memory all at once.

### 2. Map to columns

| tp column | q type | From the file | Note |
|---|---|---|---|
| `sym` | symbol | `symbol` | e.g. `` `SPY `` |
| `price` | float | `price` | decimal string → `float64` → q float |
| `size` | long | `size` | shares |
| `side` | char | `side` | aggressor: `"B"` buy, `"A"` sell, `"N"` unknown |
| `seq` | long | `sequence` | NYSE sequence number |

**No time is sent.** The file's timestamps (`ts_event`, `ts_recv`) are
dropped; the tp stamps each trade with its own current time on arrival.

### 3. Send

Each trade becomes one call on the tp:

```q
.u.upd[`trade; (enlist `SPY; enlist 778.78; enlist 38; enlist "N"; enlist 15918)]
```

Every column is a one-element vector, the same shape as a batch of rows, so
the tp can insert it straight into a table.

The connection is **kdb+ IPC over TCP**, handled by
[`github.com/sv/kdbgo`](https://github.com/sv/kdbgo):

1. **Connect once.** `kdb.DialKDBTimeout` opens a TCP connection to the tp and
   does the kdb+ handshake (credentials, then a capability byte). The same
   connection is used for the whole run, not one per message.
2. **Send asynchronously.** `conn.AsyncCall(".u.upd", ...)` sends each trade
   as an async message, like `neg[h](...)` in q: replay does not wait for a
   reply, so a slow tp never blocks the reader.
3. **Confirm at the end.** One sync call (`conn.Call("::")`) waits until the tp
   has processed every message sent before it, since q handles a connection's
   messages in order.

Each message on the wire is an 8-byte header (little-endian, async, total
length) followed by the q-encoded `(".u.upd"; `trade; columns)` list. It is not
HTTP: there are no URLs or status codes, just framed q messages on a TCP
stream.

## What the tp must provide

- A `trade` table with columns `time sym price size side seq`.
- A `.u.upd[t; x]` that puts the current time in front of the 5 columns it
  receives (`x`) before inserting/publishing, as standard `tick.q` does.

Async errors are not sent back: if `.u.upd` rejects a message (for example a
schema mismatch), replay will not know. Watch the tp console for errors.
