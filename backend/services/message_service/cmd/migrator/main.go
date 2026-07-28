package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sudo-odner/minor/backend/services/message_service/internal/migrator"
)

func main() {
	log.Printf("INFO: Starting migrator...")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbHosts, exists := os.LookupEnv("CASSANDRA_HOSTS")
	if !exists {
		log.Printf("ERROR: Migrator stoped with error: CASSANDRA_HOSTS not set")
	}
	hosts := strings.Split(dbHosts, ",")

	if err := migrator.RunMigrations(ctx, hosts, "message", 5, 5*time.Second); err != nil {
		log.Printf("ERROR: Migrator stoped with error: %v", err)
		return
	}
	log.Printf("INFO: Migrator successful created")
}
