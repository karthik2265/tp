// TODO: Add password authentication later (fh -> tp)

\l schema.q

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
  // TODO: send to rts
  }

.u.log:{[msg]
  show " | " sv ("time: ",(string .z.p);"msg: "," " sv (.Q.s1 each msg))
  }


