#!/bin/bash
# start tp, rts and rdb. Ctrl-C stops all of them.

export TP_HOST=localhost
export TP_PORT=5020

(cd tp && q main.q -p $TP_PORT) &
sleep 1
(cd rts && q main.q) &
(cd rdb && q main.q) &
wait

# Sample data to send to tp (from a q session):
#
# h:hopen `::5020
#
# one row:
# h(`.u.upd; `trade; (enlist`SPY; enlist 589.25; enlist 100; enlist"B"))
#
# multiple rows:
# h(`.u.upd; `trade; (`SPY`AAPL`SPY; 589.30 229.15 589.35; 200 75 10; "BSB"))
