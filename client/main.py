# tca tool - vwap

import os
from datetime import datetime

import pykx as kx

RDB_HOST = os.getenv("RDB_HOST", "localhost")
RDB_PORT = int(os.getenv("RDB_PORT", "5010"))
HDB_HOST = os.getenv("HDB_HOST", "localhost")
HDB_PORT = int(os.getenv("HDB_PORT", "5011"))


def vwap(
    sym: str,
    start: datetime = None,
    end: datetime = None,
    use_hdb: bool = False,
) -> float:
    """
    Volume Weighted Average Price: sum(price * size) / sum(size).

    Queries the RDB for intraday data (use_hdb=False) or the HDB for
    historical data (use_hdb=True).
    """
    host = HDB_HOST if use_hdb else RDB_HOST
    port = HDB_PORT if use_hdb else RDB_PORT

    where = _build_where(sym, start, end)

    with kx.SyncQConnection(host=host, port=port) as q:
        result = q(f"select vwap: sum[price*size]%sum size from trade where {where}")
        return float(result.pd()["vwap"].iloc[0])


def _build_where(sym: str, start: datetime = None, end: datetime = None) -> str:
    clauses = [f"sym=`{sym}"]
    if start:
        clauses.append(f"time>=`timestamp${start.strftime('%Y.%m.%dD%H:%M:%S.%f000')}")
    if end:
        clauses.append(f"time<=`timestamp${end.strftime('%Y.%m.%dD%H:%M:%S.%f000')}")
    return ", ".join(clauses)


if __name__ == "__main__":
    price = vwap("MSFT")
    print(f"MSFT VWAP: {price:.4f}")

    # historical example (uncomment to use)
    # price = vwap("SPY", start=datetime(2025, 1, 2, 9, 30), use_hdb=True)
