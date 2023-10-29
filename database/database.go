package database

import (
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type Db struct {
	Site *Site

	sqlDb       *sql.DB
	dirtyTables map[string]bool
}

// Goes through all values that come from db queries,
// reruns them if necessary.
func (db *Db) SyncValues() {
	syncVal(db.Site, db.dirtyTables)
	db.dirtyTables = make(map[string]bool)
}

// A value that comes from a db query.
// It provides methods to allow efficient syncing: see syncVal().
type valueFromQuery interface {
	dependsOnTables() []string
	refreshValue()
}

// If a table that val depends on has been updated ("dirty"),
// rerun query to update its value.
func syncVal(val valueFromQuery, dirtyTables map[string]bool) {
	shouldRefresh := false
	for _, t := range val.dependsOnTables() {
		if dirtyTables[t] == true {
			shouldRefresh = true
			break
		}
	}

	if shouldRefresh {
		val.refreshValue()
	}
}

var _ valueFromQuery = (*Site)(nil)

type Site struct {
	Title   string
	Tagline string
}

func (s *Site) dependsOnTables() []string {
	return []string{"site"}
}

func (s *Site) refreshValue() {
	row := currentDb.sqlDb.QueryRow("select title, tagline from site;")
	err := row.Scan(&s.Title, &s.Tagline)
	fmt.Println("Got:", s.Title)
	if err != nil {
		panic(err)
	}
}

var currentDb *Db

func init() {
	sql.Register(
		"sqlite3_extended",
		&sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				conn.RegisterUpdateHook(func(op int, db string, table string, rowid int64) {
					currentDb.dirtyTables[table] = true
				})
				return nil
			},
		},
	)
}

//go:embed schema.sql
var dbSchema string

func ConnectDb(path string) *Db {
	if currentDb != nil {
		err := currentDb.sqlDb.Close()
		if err != nil {
			panic(err)
		}
	}

	var db Db
	var err error
	db.sqlDb, err = sql.Open("sqlite3_extended", path)
	if err != nil {
		panic(err)
	}

	db.Site = &Site{}
	db.dirtyTables = make(map[string]bool)
	db.dirtyTables["site"] = true

	db.sqlDb.Exec("pragma foreign_keys = on;")
	db.sqlDb.Exec("pragma busy_timeout = 4000;")

	currentDb = &db

	return currentDb
}

func (db *Db) CreateSchema() *Db {
	_, err := db.sqlDb.Exec(dbSchema)
	if err != nil {
		panic(err)
	}
	return db
}

func (db *Db) RunSql(query string, args ...any) {
	_, err := db.sqlDb.Exec(query, args...)
	if err != nil {
		panic(err) // FIXME
	}
}
