// TODO: setup authentication later

/ read the TP address from the environment
host:getenv`TP_HOST;
port:getenv`TP_PORT;
if[not count host; '"TP_HOST is not set"];
if[not count port; '"TP_PORT is not set"];

/ subscribe to `SPY
h: hopen hsym `$host,":",port;
response: h(".u.sub";`trade;`SPY);

/ create a local trade table in memory
(response 0) set (response 1);

-1 "subscribed to trade for SPY on ",host,":",port;

/ get updates
upd: {[t;x]
  fn: "upd";
  .u.log[(fn;"got";count x;"row(s) for";t;"data";x)];
  t insert x;
  .u.log[(fn;"successfully inserted data into the table")];
  }

.u.log:{[msg]
  show " | " sv ("time: ",(string .z.p);"msg: "," " sv (.Q.s1 each msg))
  }