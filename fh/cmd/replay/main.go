// Command replay publishes historical trades to the kdb+ ticker plant, one
// trade per second.
//
// It reads every YYYYMMDD.json.zst file in trades_data in date order,
// starting from the earliest, and sends each trade to the ticker plant as
// .u.upd[`trade; (sym; price; size; side)], each column holding that one
// trade. The ticker plant adds the timestamp.
//
// The ticker plant address comes from the environment (or fh/.env):
// TP_HOST, TP_PORT and optionally TP_AUTH ("user:password").
//
// Usage (from fh/):
//
//	go run ./cmd/replay
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"fh/config"

	"github.com/klauspost/compress/zstd"
	kdb "github.com/sv/kdbgo"
)

const (
	dataDir  = "trades_data" // relative to fh/, where replay is run from
	table    = "trade"       // ticker plant table
	interval = time.Second   // time between trades
)

// Trade is the part of a Databento trades.json line that is published.
// The file's timestamps are not sent: the ticker plant stamps each trade
// with its own time when it arrives.
type Trade struct {
	Side   string  `json:"side"` // "A" ask (sell aggressor), "B" bid (buy aggressor), "N" none
	Price  float64 `json:"price,string"`
	Size   int64   `json:"size"`
	Symbol string  `json:"symbol"`
}

// columns returns the trade as (sym; price; size; side), each column a
// one-element vector. There is no time column: the ticker plant adds it in
// front, giving the table time sym price size side.
func columns(t *Trade) *kdb.K {
	side := "N"
	if t.Side != "" {
		side = t.Side[:1]
	}
	return kdb.NewList(
		kdb.SymbolV([]string{t.Symbol}),
		kdb.FloatV([]float64{t.Price}),
		kdb.LongV([]int64{t.Size}),
		&kdb.K{Type: kdb.KC, Attr: kdb.NONE, Data: side}, // char vector; kdbgo has no helper
	)
}

var dayFile = regexp.MustCompile(`^\d{8}\.json\.zst$`)

type replayer struct {
	conn  *kdb.KDBConn
	table *kdb.K
	tick  *time.Ticker
}

// replayFile publishes one day's trades, waiting for the next tick before
// each one.
func (r *replayer) replayFile(ctx context.Context, path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	zr, err := zstd.NewReader(f)
	if err != nil {
		return 0, err
	}
	defer zr.Close()

	sc := bufio.NewScanner(zr)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)

	var n int
	var t Trade
	for sc.Scan() {
		t = Trade{}
		if err := json.Unmarshal(sc.Bytes(), &t); err != nil {
			return n, fmt.Errorf("%s line %d: %w", path, n+1, err)
		}

		select {
		case <-r.tick.C:
		case <-ctx.Done():
			return n, ctx.Err()
		}

		if err := r.conn.AsyncCall(".u.upd", r.table, columns(&t)); err != nil {
			return n, err
		}
		n++
	}
	return n, sc.Err()
}

func main() {
	tp := config.LoadTP()
	addr := fmt.Sprintf("%s:%d", tp.Host, tp.Port)

	// File names sort chronologically, and os.ReadDir returns them sorted.
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		log.Fatal(err)
	}
	var files []string
	for _, e := range entries {
		if dayFile.MatchString(e.Name()) {
			files = append(files, filepath.Join(dataDir, e.Name()))
		}
	}
	if len(files) == 0 {
		log.Fatalf("no YYYYMMDD.json.zst files in %s", dataDir)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := kdb.DialKDBTimeout(tp.Host, tp.Port, tp.Auth, 5*time.Second)
	if err != nil {
		log.Fatalf("connect to ticker plant %s: %v", addr, err)
	}
	defer conn.Close()

	r := &replayer{conn: conn, table: kdb.Symbol(table), tick: time.NewTicker(interval)}
	defer r.tick.Stop()
	log.Printf("replaying %d days to %s (%s .. %s), one trade every %s",
		len(files), addr, filepath.Base(files[0]), filepath.Base(files[len(files)-1]), interval)

	total := 0
	for _, path := range files {
		n, err := r.replayFile(ctx, path)
		total += n
		if err != nil {
			if ctx.Err() != nil {
				log.Printf("interrupted during %s after %d trades", filepath.Base(path), n)
				break
			}
			log.Fatal(err)
		}
		log.Printf("%s: %d trades", filepath.Base(path), n)
	}

	// A sync round trip returns only after the ticker plant has processed every
	// async message sent before it. Errors inside .u.upd are not reported back
	// (async errors stay on the ticker plant), so check its console for those.
	if _, err := conn.Call("::"); err != nil {
		log.Fatalf("ticker plant: %v", err)
	}
	log.Printf("done: %d trades sent", total)
}
