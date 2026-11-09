package main

import (
	"context"
	"fh/config"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata/stream"
)

func main() {
	config := config.LoadConfig()

	if config.APIKey == "" || config.APISecret == "" {
		log.Fatal("Missing Alpaca API credentials")
	}

	// Handle graceful shutdown.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Handler for incoming stock trades.
	onTrade := func(t stream.Trade) {
		fmt.Printf(
			"Symbol: %-5s Price: %-10.2f Size: %-6d Exchange: %-5s Time: %s\n",
			t.Symbol,
			t.Price,
			t.Size,
			t.Exchange,
			t.Timestamp,
		)
	}

	// Create a stock market data stream.
	client := stream.NewStocksClient(
		marketdata.IEX,
		stream.WithCredentials(config.APIKey, config.APISecret),
		stream.WithReconnectSettings(0, 2*time.Second),
		stream.WithLogger(stream.DefaultLogger()),
	)

	// Establish WebSocket connection.
	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Connection failed: %v", err)
	}

	log.Println("Connected to Alpaca market data stream")

	symbols := []string{
		"UBS",   // UBS
		"JPM",   // JPMorgan Chase
		"GS",    // Goldman Sachs
		"MS",    // Morgan Stanley
		"C",     // Citigroup
		"BAC",   // Bank of America
		"BCS",   // Barclays
		"DB",    // Deutsche Bank
		"BNPQY", // BNP Paribas ADR
	}

	// Subscribe to trade events.
	if err := client.SubscribeToTrades(
		onTrade,
		symbols...,
	); err != nil {
		log.Fatalf("Subscription failed: %v", err)
	}

	log.Println("Subscribed to trade events for UBS and its competitors")

	// Wait for shutdown or stream termination.
	select {
	case <-ctx.Done():
		log.Println("Shutting down")
	case err, ok := <-client.Terminated():
		if ok && err != nil {
			log.Printf("Stream terminated: %v", err)
		}
	}
}
