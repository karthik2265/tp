// TODO: Add password authentication later (fh -> tp)

\l schema.q

.u.t: tables`.;
// set up watchers/subscribers
.u.w: .u.t!(count .u.t)#();

// recieve the data from feed-handler
// log to the file
// send updates to rts
.u.upd: {[t; x]
  fn: ".u.upd";
  .u.log[(fn;t;x)];
  // validate the data
  if[not t in tables`.; '"unknow table: ", .Q.s1 t];
  if[not (-1+count cols t)=count x; '"expected ", string[-1+count cols t], " columns, got ", string count x];
  // add timestamp
  x: (enlist (count first x)#.z.p), x;
  // send to rts
  newRows: flip (cols t)!x;
  .u.pub[t; newRows]
  }

.u.pub: {[t; rows]
  fn: ".u.pub";
  .u.log[(fn;"publising result for table:";t;"and data:";rows)];
  / publish to watchers
  watchers: .u.w[t];
  {[t;rows;watcher]
    h: watcher 0;
    s: watcher 1;
    d: $[s~`;rows;select from rows where sym in s];
    if[count d; neg[h] (`upd; t; d); .u.log[("successfully published to";watcher;"to table";t;"and data";rows)]];
    }[t;rows] each watchers;
  }

.u.log:{[msg]
  show " | " sv ("time: ",(string .z.p);"msg: "," " sv (.Q.s1 each msg))
  }

.u.sub: {[t;s]
  fn: ".u.sub";
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


