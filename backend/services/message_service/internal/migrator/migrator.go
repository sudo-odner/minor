package migrator

import (
	"context"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/migrate"
)

//go:embed migrations/*.cql
var migrationFiles embed.FS

func reconnect(ctx context.Context, clusterCfg *gocql.ClusterConfig, maxRetries int, retryInterval time.Duration) (*gocql.Session, error) {
	const op = "migrator.reconnect"

	ticker := time.NewTicker(retryInterval)
	defer ticker.Stop()

	for try := 1; try <= maxRetries; try++ {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%s: canceled by context: %w", op, ctx.Err())
		default:
		}

		session, err := clusterCfg.CreateSession()
		if err == nil {
			return session, nil
		}

		log.Printf("%s: database not answering (try %d/%d)", op, try, maxRetries)
		if try == maxRetries {
			break
		}

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%s: canceled by context: %w", op, ctx.Err())
		case <-ticker.C:
		}
	}
	return nil, fmt.Errorf("%s: connection limit exceeded (%d attempts)", op, maxRetries)
}

// TODO: Better create config settings for migrations (for setting keyspace)
func RunMigrations(ctx context.Context, hosts []string, keyspace string, maxRetries int, retryInterval time.Duration) error {
	const op = "migrator.RunMigrations"
	// Init cluster config
	cluster := gocql.NewCluster(hosts...)
	cluster.Timeout = 5 * time.Second

	// Connect to cluster cassandra and create keyspcase
	err := func() error {
		session, err := reconnect(ctx, cluster, maxRetries, retryInterval)
		if err != nil {
			return err
		}
		defer session.Close()

		log.Printf("%s: success connect to Cassandra", op)

		createKeySpaceQuery := fmt.Sprintf(`
			CREATE KEYSPACE IF NOT EXISTS %s
			WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
		`, keyspace)
		if err := session.Query(createKeySpaceQuery).Exec(); err != nil {
			return fmt.Errorf("failed create keyspace: %w", err)
		}
		log.Printf("%s: keyspace '%s' is created/checked", op, keyspace)
		return nil
	}()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Connect to cluser with keyspace and run migrations
	cluster.Keyspace = keyspace
	cqlxSession, err := gocqlx.WrapSession(reconnect(ctx, cluster, maxRetries, retryInterval))
	if err != nil {
		return fmt.Errorf("%s: failed connect to keyspace: %w", op, err)
	}
	defer cqlxSession.Close()

	log.Printf("%s: running migration on keyspace '%s'...", op, keyspace)
	if err := migrate.FromFS(ctx, cqlxSession, migrationFiles); err != nil {
		return fmt.Errorf("%s: falied migrate: %w", op, err)
	}
	log.Printf("%s: database schema successufully update", op)
	return nil
}
