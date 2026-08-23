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
