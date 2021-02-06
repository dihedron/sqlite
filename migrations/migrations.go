package migrations

import "embed"

//go:embed *.sql
// Migrations contains all SQL files used to initialise the SQLIte3 database.
// It uses Go 1.16 embed directive to embed into the application all SQL
// scripts in this directory.
var Migrations embed.FS
