# client

TCA (Transaction Cost Analysis) tool for querying the ticker plant's RDB and HDB.

Currently implements **VWAP** (Volume Weighted Average Price).

## Setup

```bash
pip install -r requirements.txt
```

> `pykx` is KX's official Python library. It works in **unlicensed / IPC-only mode**
> without a kdb+ license — which is all we need here (pure query client).
> Set `PYKX_UNLICENSED=true` to force unlicensed mode explicitly.

## Usage

```python
from main import vwap

# intraday — queries the RDB (default)
price = vwap("SPY")

# historical — queries the HDB
from datetime import datetime
price = vwap("SPY", start=datetime(2025, 1, 2, 9, 30), use_hdb=True)
```

Or run directly:

```bash
python main.py
```

## Environment variables

| Variable   | Default     | Description          |
|------------|-------------|----------------------|
| `RDB_HOST` | `localhost` | RDB host             |
| `RDB_PORT` | `5010`      | RDB port             |
| `HDB_HOST` | `localhost` | HDB host             |
| `HDB_PORT` | `5011`      | HDB port             |

## Data source

Queries the `trade` table (schema: `time sym price size side`).

- **RDB** (port 5010) — in-memory, today's data only
- **HDB** (port 5011) — on-disk partitioned database, historical data
