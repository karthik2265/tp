// TODO: setup authentication later

.u.t: tables`.;
/ read the TP address from the environment
host:getenv`TP_HOST;
port:getenv`TP_PORT;
if[not count host; '"TP_HOST is not set"];
if[not count port; '"TP_PORT is not set"];

/ subscribe to all symbols
h: hopen hsym `$host,":",port;
response: h(".u.sub";`trade;`);

/ create a local trade table in memory
(response 0) set (response 1);

-1 "subscribed to trades on ",host,":",port;

/ get updates
upd: {[t;d]
  fn: "rdb.u.upd";
  .u.log[(fn;"got";count d;"row(s) for";t;"data";d)];
  t insert d;
  .u.log[(fn;"successfully inserted data into the table")];
  }

/ save to hdb at the eod
.u.end: {[dt]
  fn: "rdb.u.end";
  .u.log[(fn;"saving data to hdb at the end of the day for date:";d)];
  r: `:../hdb/data;
  {[fn;r;dt;t]
    p: ` sv root,(`$string dt),t,`;
    .u.log[(fn;"saving";t;"to";p)];
    p set .Q.en[r; `sym xasc value t];
    @[p; `sym; `p#];
    delete from t;
    }[fn;r;dt] each .u.t;
  }

.u.log:{[msg]
  show " | " sv ("time: ",(string .z.p);"msg: "," " sv (.Q.s1 each msg))
  }