package cassandra

import (
	"fmt"

	"github.com/gocql/gocql"
	"github.com/sudo-odner/minor/backend/services/chat_service/internal/config"
)

type Repository struct {
	session *gocql.Session
}

func New(cfg *config.Cassandra) (*Repository, error) {
	const op = "repository.cassandra.New"

	cluster := gocql.NewCluster(cfg.Host)
	cluster.Keyspace = cfg.Keyspace
	if cfg.Username != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{
			Username: cfg.Username,
			Password: cfg.Password,
		}
	}

	cluster.Timeout = cfg.Timeout
	consistency, err := gocql.ParseConsistencyWrapper(cfg.Consistency)
	if err != nil {
		return nil, fmt.Errorf("%s: unncorect consistency type %w", op, err)
	}
	cluster.Consistency = consistency
	cluster.NumConns = cfg.NumConns

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create session: %w", op, err)
	}
	return &Repository{
		session: session,
	}, nil
}

func (r *Repository) Ping() error {
	return r.session.Query("SELECT now() FROM system.local").Exec()
}

func (r *Repository) Close() {
	if r.session != nil {
		r.session.Close()
	}
}
