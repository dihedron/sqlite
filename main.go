package main

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/dihedron/sqlite/log"
	"github.com/dihedron/sqlite/migrations"
	"github.com/dihedron/sqlite/sqlite"
	"go.uber.org/zap"
)

func main() {

	defer log.L.Sync()

	var scripts fs.FS
	if len(os.Args) == 1 {
		scripts = migrations.Migrations
	} else {
		scripts = os.DirFS(os.Args[1])
	}

	db, err := sqlite.InitDB("database/sqlite3.db", scripts)
	if err != nil {
		log.L.Error("error opening database", zap.Error(err))
		os.Exit(1)
	}

	count := 0
	row := db.QueryRow("SELECT count(*) FROM pairs")
	if err := row.Scan(&count); err != nil {
		log.L.Error("error querying database", zap.Error(err))
		os.Exit(1)
	}
	fmt.Printf("count: %d\n", count)
}
