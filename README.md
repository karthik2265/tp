# This is a vannila ticker plant setup

client: send queries to hdb/rdb to get the data needed.
fd: feed handler - send the tick data to tp.
hdb: historical database - store all of the data recieved from feedhandler.
rdb: real time database - store todays data in memory.
rts: real time subscriber - get the real time data.
tp: ticker plant - get the data from from fd and send to rbd and rts.