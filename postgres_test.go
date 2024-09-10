package postgres

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// On Ubuntu, it's necessary to be sudo to access docker socket.
// Run the following to ensure sudo has access to go:
//
// sudo env PATH=$PATH go test -v -run=TestPostgres

var testParams = PostgresOpts{
	Host:     "localhost",
	User:     "mypostgresuser",
	Password: "mysecretpassword",
	Name:     "mydatabase",
	Sslmode:  "disable",
}

func TestPing(t *testing.T) {
	ctx := context.Background()
	db, teardown, err := NewTestDB(ctx, testParams)
	require.NoError(t, err)
	defer teardown()
	t.Logf("Postgres is running on %s:%s", db.Host, db.Port)

	err = db.Ping(ctx)
	require.NoError(t, err)
	t.Log("Ping successful!")
}

func TestQuery(t *testing.T) {
	ctx := context.Background()
	db, teardown, err := NewTestDB(ctx, testParams)
	require.NoError(t, err)
	defer teardown()
	t.Logf("Postgres is running on %s:%s", db.Host, db.Port)

	col := "datname"
	table := "pg_database"
	query := fmt.Sprintf("SELECT %s FROM %s", col, table)
	t.Logf("running query:\n  %s", query)
	results, err := db.Query(ctx, query)
	require.NoError(t, err)
	for _, result := range results {
		value, ok := result[col].([]uint8)
		if !ok {
			t.Errorf("expected []uint8 for column %s, got %T", col, result[col])
			continue
		}
		t.Logf("result: %v=%v",
			col,
			string(value),
		)
	}
}
