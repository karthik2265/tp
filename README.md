# This is a vannila ticker plant setup

client: send queries to hdb/rdb to get the data needed.
fd: feed handler - send the tick data to tp.
hdb: historical database - store all of the data recieved from feedhandler.
rdb: real time database - store todays data in memory.
rts: real time subscriber - get the real time data.
tp: ticker plant - get the data from from fd and send to rbd and rts.
## Replaying historical trades

From the repo root, in three terminals:

```sh
q tp/tp.q -p 5020      # ticker plant (new log in tp/logs each start)
q rdb/rdb.q -p 5021    # rdb: subscribes to the tp, loads rdb/tca.q

cd fh
go run ./cmd/rename    # once: trades_data/xnys-pillar-20261012.trades.json.zst -> 20261012.json.zst
go run ./cmd/replay    # all days from the earliest; -speed 1 for real time, -from/-to YYYYMMDD
```

TCA (on the rdb, times in UTC): `.tca.vwap`, `.tca.twap`, `.tca.pwp`, `.tca.arrival`, `.tca.report` — see `rdb/tca.q`.
