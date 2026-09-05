package message

import "embed"

//go:embed migrations/*.cql
var MigrationsFS embed.FS
