/ bind to port passed via -p flag (default 5011)
if[not system"p"; system"p 5011"];

/ load the on-disk partitioned database
\l data

.u.log:{[msg]
  show " | " sv ("time: ",(string .z.p);"msg: "," " sv (.Q.s1 each msg))
  }

-1 "hdb ready on port ",string system"p";
