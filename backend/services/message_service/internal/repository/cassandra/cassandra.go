package cassandra

import (
	"github.com/gocql/gocql"
)

type Repository struct {
	session *gocql.Session
}

func New(session *gocql.Session) *Repository {
	return &Repository{
		session: session,
	}
}

func (r *Repository) Ping() error {
	return r.session.Query("SELECT now() FROM system.local").Exec()
}

func (r *Repository) Close() error {
	if r.session != nil {
		r.session.Close()
	}
	return nil
}
