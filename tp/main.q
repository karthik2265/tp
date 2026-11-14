// TODO: Add password authentication later (fh -> tp)

\l schema.q

.u.t: tables`.;

// set up watchers/subscribers
.u.w: .u.t!(count .u.t)#();

// recieve the data from feed-handler
// log to the file
// send updates to rts
.u.upd: {[t; d]
  fn: "tp.u.upd";
  .u.log[(fn;t;d)];
  // validate the data
  if[not t in tables`.; '"unknow table: ", .Q.s1 t];
  if[not (-1+count cols t)=count d; '"expected ", string[-1+count cols t], " columns, got ", string count d];
  // add timestamp
  d: (enlist (count first d)#.z.p), d;
  // send to rts
  d: flip (cols t)!d;
  .u.pub[t; d]
  }

.u.pub: {[t; d]
  fn: "tp.u.pub";
  .u.log[(fn;"publising result for table:";t;"and data:";d)];
  / publish to watchers
  ws: .u.w[t];
  {[t;d;w]
    h: w 0;
    s: w 1;
    d: $[s~`;d;select from d where sym in s];
    if[count d; neg[h] (`upd; t; d); .u.log[("successfully published to";w;"to table";t;"and data";d)]];
    }[t;d] each ws;
  }

// TODO: save to log file.
.u.log:{[msg]
  show " | " sv ("time: ",(string .z.p);"msg: "," " sv (.Q.s1 each msg))
  }

.u.sub: {[t;s]
  fn: "tp.u.sub";
  .u.log[(fn;"subscription request recieved for";t;"and symbols";s)];
  if[not t in .u.t; '"unknown table"];
  // remove old watchers
  .u.w[t]: .u.w[t][where not (.z.w=(first each .u.w[t]))];
  // add the new watcher
  .u.w[t] ,: enlist (.z.w;s);
  .u.log[(fn;"subscription setup successfully")];
  // return table name and schema
  :(t; value t)
  }

// check eod
.u.d: .z.d;
.z.ts: {[x] if[.z.d > .u.d; .u.eod[]]; }
\t 1000

.u.eod: {[]
  / every handle subscribed to any table, once each
  hs: distinct first each raze value .u.w;
  / tell each one that day .u.d is finished
  {[d;h] neg[h] (`.u.end; d)}[.u.d] each hs;
  .u.d +: 1;
  }
