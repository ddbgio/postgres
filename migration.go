package postgres

import (
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
)

// Migrations represents a single SQL migration file,
// including the direction, file name, and content
type Migration struct {
	Direction string
	Filename  string
	Content   string
}

// Migrations reads the migrations directory for files matching
// the direction, returning an error if no files are found or
// returning a list of Migration objects containing the file name and contents
func Migrations(migrationsDir string, direction string) ([]Migration, error) {
	switch direction {
	case "up":
		slog.Debug("fetching up migrations", "dir", migrationsDir)
	case "down":
		slog.Debug("fetching down migrations", "dir", migrationsDir)
	default:
		return nil, fmt.Errorf(
			"invalid direction '%s', must be 'up' or 'down'",
			direction,
		)
	}

	// read directory
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("unable to read directory '%s': %w",
			migrationsDir,
			err,
		)
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no files found in directory '%s'", migrationsDir)
	}

	// capture data
	var migrations []Migration
	for _, file := range files {
		if file.IsDir() {
			slog.Warn("skipping directory", "name", file.Name())
			continue
		}
		expectedEnding := fmt.Sprintf("%s.sql", direction)
		if strings.Contains(file.Name(), expectedEnding) {
			bytes, err := os.ReadFile(path.Join(migrationsDir, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("unable to read file '%s': %w", file.Name(), err)
			}
			if len(bytes) == 0 {
				return nil, fmt.Errorf("empty file '%s'", file.Name())
			}
			migration := Migration{
				Direction: direction,
				Filename:  file.Name(),
				Content:   string(bytes),
			}
			migrations = append(migrations, migration)
		}
	}
	if len(migrations) == 0 {
		return nil, fmt.Errorf("no files found with ending '%s.sql'", direction)
	}

	return migrations, nil
}
