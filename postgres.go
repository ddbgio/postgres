package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const portDefault = "5432"

// Postgres represents an instantiated postgres database
type Postgres struct {
	Host     string
	User     string
	Password string
	Name     string
	Sslmode  string
	Port     string
	Conn     *sql.DB
}

// PostgresOpts define settings for a new database
type PostgresOpts struct {
	Host     string
	User     string
	Password string
	Name     string
	Sslmode  string
}

// ConnStr formats a connection string in either the key-value or URI format
func (db *Postgres) ConnStr(uriFormat bool) (string, error) {
	if db.Host == "" {
		slog.Info("no host provided, defaulting to localhost")
		db.Host = "localhost"
	}
	if db.Password == "" {
		return "", fmt.Errorf("password required")
	}
	if db.Port == "" {
		db.Port = portDefault
	}
	// TODO additional validation?
	if db.Sslmode == "disable" {
		slog.Warn("running SSL_mode=disable")
	}
	var connStr string
	if uriFormat {
		// URI format
		sslParam := ""
		if db.Sslmode == "disable" {
			sslParam = "?sslmode=disable"
		}
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s%s",
			db.User, db.Password, db.Host, db.Port, db.Name, sslParam,
		)
	} else {
		// key value format
		connStr = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s sslmode=%s port=%s",
			db.Host, db.User, db.Password, db.Name, db.Sslmode, db.Port,
		)
	}

	return connStr, nil
}

func (db *Postgres) Pool(ctx context.Context) (*pgxpool.Pool, error) {
	connStr, err := db.ConnStr(true)
	if err != nil {
		return nil, fmt.Errorf("unable format connection string: %w", err)
	}
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string")
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool")
	}
	return pool, nil
}

// Open
func (db *Postgres) Open() error {
	connStr, err := db.ConnStr(true)
	if err != nil {
		return fmt.Errorf("unable to create connection string: %w", err)
	}
	database, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	db.Conn = database
	return nil
}

func (db *Postgres) Ping(ctx context.Context) error {
	err := db.Conn.Ping()
	if err != nil {
		return fmt.Errorf("unable to ping db: %w", err)
	}
	// slog.Debug("ping successful")
	return nil
}
func (db *Postgres) Execute(ctx context.Context, query string, vars ...any) (int64, int64, error) {
	return 0, 0, fmt.Errorf("not implemented, use Query instead")
}

func (db *Postgres) Query(ctx context.Context, query string, vars ...any) ([]map[string]any, error) {
	// TODO test vars, probably wrong
	rows, err := db.Conn.QueryContext(ctx, query, vars...)
	if err != nil {
		return nil, fmt.Errorf("unable to query db: %w", err)
	}
	defer rows.Close()
	// slog.Debug("getting results")

	colNames, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("unable to get columns: %w", err)
	}
	var results []map[string]any

	width := len(colNames)
	for rows.Next() {
		vals := make([]any, width)
		valPtrs := make([]any, width)
		for i := range vals {
			valPtrs[i] = &vals[i]
		}
		if err := rows.Scan(valPtrs...); err != nil {
			return nil, fmt.Errorf("unable to scan rows: %w", err)
		}
		result := make(map[string]any)
		for i, name := range colNames {
			result[name] = vals[i]
			slog.Debug("added result", "key", name, "value", result[name])
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("unable to get rows: %w", err)
	}
	return results, nil
}

func (db *Postgres) Close() error {
	err := db.Conn.Close()
	if err != nil {
		return fmt.Errorf("problem when closing db connection: %w", err)
	}
	return nil
}

func NewDB(ctx context.Context, opts PostgresOpts) (Postgres, error) {
	db := Postgres{
		Host:     opts.Host,
		User:     opts.User,
		Password: opts.Password,
		Name:     opts.Name,
		Sslmode:  opts.Sslmode,
	}
	err := db.Open()
	if err != nil {
		return Postgres{}, fmt.Errorf("unable to open db: %w", err)
	}
	return db, nil
}

func NewTestDB(ctx context.Context, opts PostgresOpts) (Postgres, func(), error) {
	var db = Postgres{
		Host:     opts.Host,
		User:     opts.User,
		Password: opts.Password,
		Name:     opts.Name,
		Sslmode:  opts.Sslmode,
	}
	postgresContainer, err := postgres.Run(ctx,
		"docker.io/postgres:16-alpine",
		// postgres.WithInitScripts(filepath.Join("testdata", "init-user-db.sh")),
		// postgres.WithConfigFile(filepath.Join("testdata", "my-postgres.conf")),
		postgres.WithDatabase(db.Name),
		postgres.WithUsername(db.User),
		postgres.WithPassword(db.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return Postgres{}, nil, fmt.Errorf("run failed: %w", err)
	}
	teardown := func() {
		db.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			log.Fatalf("failed to terminate container: %s", err)
		}
	}

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		return Postgres{}, teardown, fmt.Errorf("unable to get host: %w", err)
	}
	db.Host = host

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return Postgres{}, teardown, fmt.Errorf("unable to get port: %w", err)
	}
	db.Port = port.Port()
	err = db.Open()
	if err != nil {
		return Postgres{}, teardown, fmt.Errorf("unable to open db: %w", err)
	}
	return db, teardown, nil
}
