// Custom dbmate driver for modernc.org/sqlite (CGO-free)
package sqlite

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	"github.com/amacneil/dbmate/v2/pkg/dbutil"
)

func init() {
	// Register our custom SQLite driver with dbmate
	dbmate.RegisterDriver(NewModerncDriver, "sqlite")
	dbmate.RegisterDriver(NewModerncDriver, "sqlite3")
}

// ModerncDriver provides dbmate functionality for modernc.org/sqlite
type ModerncDriver struct {
	migrationsTableName string
	databaseURL         *url.URL
	log                 io.Writer
}

// NewModerncDriver initializes the driver
func NewModerncDriver(config dbmate.DriverConfig) dbmate.Driver {
	return &ModerncDriver{
		migrationsTableName: config.MigrationsTableName,
		databaseURL:         config.DatabaseURL,
		log:                 config.Log,
	}
}

// ConnectionString converts a URL into a valid connection string
func ConnectionString(u *url.URL) string {
	newURL := *u
	newURL.Scheme = ""

	if newURL.Opaque == "" && newURL.Path != "" {
		newURL.Opaque = "//" + newURL.Host + dbutil.MustUnescapePath(newURL.Path)
		newURL.Path = ""
	}

	str := regexp.MustCompile("^//+").ReplaceAllString(newURL.String(), "/")
	return str
}

// Open creates a new database connection
func (drv *ModerncDriver) Open() (*sql.DB, error) {
	return sql.Open("sqlite", ConnectionString(drv.databaseURL))
}

// CreateDatabase creates the specified database
func (drv *ModerncDriver) CreateDatabase() error {
	fmt.Fprintf(drv.log, "Creating: %s\n", ConnectionString(drv.databaseURL))

	db, err := drv.Open()
	if err != nil {
		return err
	}
	defer dbutil.MustClose(db)

	return db.Ping()
}

// DropDatabase drops the specified database (if it exists)
func (drv *ModerncDriver) DropDatabase() error {
	path := ConnectionString(drv.databaseURL)
	fmt.Fprintf(drv.log, "Dropping: %s\n", path)

	exists, err := drv.DatabaseExists()
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	return os.Remove(path)
}

func (drv *ModerncDriver) schemaMigrationsDump(db *sql.DB) ([]byte, error) {
	migrationsTable := drv.quotedMigrationsTableName()

	migrations, err := dbutil.QueryColumn(db,
		fmt.Sprintf("select quote(version) from %s order by version asc", migrationsTable))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	buf.WriteString("-- Dbmate schema migrations\n")

	if len(migrations) > 0 {
		buf.WriteString(
			fmt.Sprintf("INSERT INTO %s (version) VALUES\n  (", migrationsTable) +
				strings.Join(migrations, "),\n  (") +
				");\n")
	}

	return buf.Bytes(), nil
}

// DumpSchema returns the current database schema
func (drv *ModerncDriver) DumpSchema(db *sql.DB) ([]byte, error) {
	path := ConnectionString(drv.databaseURL)
	schema, err := dbutil.RunCommand("sqlite3", path, ".schema --nosys")
	if err != nil {
		return nil, err
	}

	migrations, err := drv.schemaMigrationsDump(db)
	if err != nil {
		return nil, err
	}

	schema = append(schema, migrations...)
	return dbutil.TrimLeadingSQLComments(schema)
}

// DatabaseExists determines whether the database exists
func (drv *ModerncDriver) DatabaseExists() (bool, error) {
	_, err := os.Stat(ConnectionString(drv.databaseURL))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// MigrationsTableExists checks if the schema_migrations table exists
func (drv *ModerncDriver) MigrationsTableExists(db *sql.DB) (bool, error) {
	exists := false
	err := db.QueryRow("SELECT 1 FROM sqlite_master "+
		"WHERE type='table' AND name=$1",
		drv.migrationsTableName).
		Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}

	return exists, err
}

// CreateMigrationsTable creates the schema migrations table
func (drv *ModerncDriver) CreateMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(fmt.Sprintf(
		"create table if not exists %s (version varchar(128) primary key)",
		drv.quotedMigrationsTableName()))

	return err
}

// SelectMigrations returns a list of applied migrations
func (drv *ModerncDriver) SelectMigrations(db *sql.DB, limit int) (map[string]bool, error) {
	query := fmt.Sprintf("select version from %s order by version desc", drv.quotedMigrationsTableName())
	if limit >= 0 {
		query = fmt.Sprintf("%s limit %d", query, limit)
	}
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	defer dbutil.MustClose(rows)

	migrations := map[string]bool{}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}

		migrations[version] = true
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return migrations, nil
}

// InsertMigration adds a new migration record
func (drv *ModerncDriver) InsertMigration(db dbutil.Transaction, version string) error {
	_, err := db.Exec(
		fmt.Sprintf("insert into %s (version) values (?)", drv.quotedMigrationsTableName()),
		version)

	return err
}

// DeleteMigration removes a migration record
func (drv *ModerncDriver) DeleteMigration(db dbutil.Transaction, version string) error {
	_, err := db.Exec(
		fmt.Sprintf("delete from %s where version = ?", drv.quotedMigrationsTableName()),
		version)

	return err
}

// Ping verifies a connection to the database
func (drv *ModerncDriver) Ping() error {
	db, err := drv.Open()
	if err != nil {
		return err
	}
	defer dbutil.MustClose(db)

	return db.Ping()
}

// QueryError returns a normalized version of the driver-specific error type
func (drv *ModerncDriver) QueryError(query string, err error) error {
	return &dbmate.QueryError{Err: err, Query: query}
}

func (drv *ModerncDriver) quotedMigrationsTableName() string {
	return drv.quoteIdentifier(drv.migrationsTableName)
}

// quoteIdentifier quotes a table or column name using SQLite double quotes
func (drv *ModerncDriver) quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}