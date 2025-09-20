package main

import (
	"log"
	"time"
	"os"
)

func main() {
	interval := getenv("MIGRATOR_POLL_INTERVAL", "30s")
	d, _ := time.ParseDuration(interval)
	log.Printf("migrator running, poll=%s (placeholder)", d)
	for {
		time.Sleep(d)
		log.Println("polling for schema changes (placeholder)")
	}
}

func getenv(k, d string) string { if v := os.Getenv(k); v != "" { return v }; return d }
