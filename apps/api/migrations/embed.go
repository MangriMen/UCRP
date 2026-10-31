package migrations

import "embed"

//go:embed common/*.sql postgres/*.sql
var FS embed.FS
