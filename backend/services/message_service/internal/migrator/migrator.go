package migrator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/cassandra"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	message "github.com/sudo-odner/minor/backend/services/message_service"
)

// TODO: Better create config settings for migrations (for setting keyspace)

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

		log.Printf("INFO: %s: success connect to Cassandra", op)

		createKeySpaceQuery := fmt.Sprintf(`
			CREATE KEYSPACE IF NOT EXISTS %s
			WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1};
		`, keyspace)
		if err = session.Query(createKeySpaceQuery).Exec(); err != nil {
			return fmt.Errorf("failed create keyspace: %w", err)
		}
		log.Printf("INFO: %s: keyspace '%s' is created/checked", op, keyspace)
		return nil
	}()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// Connect to cluster cassandra with keyspace
	cluster.Keyspace = keyspace
	session, err := reconnect(ctx, cluster, maxRetries, retryInterval)
	if err != nil {
		return fmt.Errorf("%s: connect to keyspace failed: %w", op, err)
	}
	defer session.Close()
	log.Printf("INFO: %s: success connect to Cassandra with keyspcae '%s'", op, keyspace)

	// Setup migrator and run Migrations
	sourceDriver, err := iofs.New(message.MigrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("%s: init source driver falied: %w", op, err)
	}

	dbDriver, err := cassandra.WithInstance(session, &cassandra.Config{
		KeyspaceName: keyspace,
	})
	if err != nil {
		return fmt.Errorf("%s: init db driver falied: %w", op, err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "cassandra", dbDriver)
	if err != nil {
		return fmt.Errorf("%s: init migrator falied: %w", op, err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("%s: run up falied: %w", op, err)
	}
	log.Printf("INFO: %s: database schema successufully update", op)
	return nil
}
