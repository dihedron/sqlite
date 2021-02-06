package migrations

import "embed"

//go:embed *.sql
// Migrations contains all SQL files used to initialise
// the SQLIte3 database.
var Migrations embed.FS
