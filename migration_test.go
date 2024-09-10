package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFetchFiles(t *testing.T) {
	migrations, err := Migrations("example-migrations", "down")
	require.NoError(t, err)
	require.NotNil(t, migrations)
	for _, migration := range migrations {
		t.Logf("migration: %v (%v)\n%v\n",
			migration.Filename,
			migration.Direction,
			migration.Content)
	}
}
