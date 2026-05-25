package migrator

import (
	"context"
	"embed"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

//go:embed migrations/*.cql
var migrationFiles embed.FS

func reconnect(ctx context.Context, clusterCfg *gocql.ClusterConfig, maxReconect int) (*gocql.Session, error) {
	const op = "migrator.reconnect"
	var systemSession *gocql.Session
	var err error

	systemSession, err = clusterCfg.CreateSession()
	if err == nil {
		return systemSession, err
	}

	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for i := 0; i < maxReconect; i++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%s: migrator cancel by context: %w", op, ctx.Err())
		case <-ticker.C:
			systemSession, err = clusterCfg.CreateSession()
			if err == nil {
				return systemSession, nil
			}
			log.Printf("%s: database not answer (try %d/%d)", op, i, maxReconect)
		}
	}
	return nil, fmt.Errorf("%s: connection attempt limit exceeded: %w", op, err)
}

func RunMigrations(ctx context.Context, hosts []string, keyspace string) error {
	const op = "migrator.RunMigrations"

	// 0. Start setting cluster
	cluster := gocql.NewCluster(hosts...)
	cluster.Consistency = gocql.Quorum
	cluster.Timeout = 5 * time.Second

	// 1. Connect to Database
	systemSession, err := reconnect(ctx, cluster, 15)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer systemSession.Close()
	log.Printf("%s: success connect to Cassandra", op)

	// 2. Creating keyspace
	createKeySpace := fmt.Sprintf(`
		CREATE KEYSPACE IF NOT EXISTS %s
		WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
	`, keyspace)

	if err := systemSession.Query(createKeySpace).Exec(); err != nil {
		return fmt.Errorf("%s: failed create keyspace: %w", op, err)
	}
	systemSession.Close()
	log.Printf("%s: keyspace '%s' is created", op, keyspace)

	// 3. Migrate migrations
	cluster.Keyspace = keyspace
	appSession, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("%s: failed connect to keyspace: %w", op, err)
	}
	defer appSession.Close()

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("%s: falied to read the embedded migrations folder: %w", op, err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cql") {
			continue
		}

		content, err := migrationFiles.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("%s: falied to read '%s' file: %w", op, entry.Name(), err)
		}
		log.Printf("%s: exec file %s", op, entry.Name())

		// Cassandra не умеет выполнять пачку DDL-запросов за раз
		// Разбиваем запросы по ';' и поледовательно выполняем запросы
		for query := range strings.SplitSeq(string(content), ";") {
			clearQuery := strings.TrimSpace(query)
			if clearQuery == "" {
				continue
			}

			if err := appSession.Query(clearQuery).Exec(); err != nil {
				return fmt.Errorf("%s: error in file '%s' on query [%s]: %w", op, entry.Name(), query, err)
			}
		}
	}
	log.Printf("%s: database schema successful complete", op)
	return nil
}
