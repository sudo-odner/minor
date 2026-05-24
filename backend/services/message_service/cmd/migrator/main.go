package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sudo-odner/minor/backend/services/message_service/internal/migrator"
)

func main() {
	const op = "cmd.migrator"

	log.Printf("[%s] Starting migrator...", op)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbHosts := os.Getenv("CASSANDRA_HOSTS")
	if dbHosts == "" {
		dbHosts = "127.0.0.1"
	}
	hosts := strings.Split(dbHosts, ",")

	if err := migrator.RunMigrations(ctx, hosts, "chat"); err != nil {
		log.Printf("[%s] Migrator stoped with error: %v", op, err)
		return
	}
	log.Printf("[%s] Migrator successful created", op)
}
