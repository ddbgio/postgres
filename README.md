# postgres
go wrappers and testcontainers for postgres

## migrations
Files for initailizing (and tearing down) a database are included in the `Migrations()` function. Migrations files should be numbered in the order they are to be run, first `down`, then `up`. As such, all `down` migrations should be odd; all `up` migrations should be even. Migrations will return a list of `Migration` objects, containing _Direction_, 

```go
// Migrations represents a single SQL migration file,
// including the direction, file name, and content
type Migration struct {
	Direction string
	Filename  string
	Content   string
}

var migrations []Migration
var err error
migrations, err = Migrations("example-migrations", "down")
if err != nil {
  return nil, fmt.Errorf("unable to fetch migrations: %w", err)
}

for _, migration := range migrations {
  slog.Info("running a migration", "direction", migration.Direction, "filename", migration.Filename)
  err := db.Execute(migration.Content, nil)
  if err != nil {
    return nil, fmt.Errorf("unable to execute migration: %w", err) 
  }
}
```

## 🏗️ testcontainers
[Testcontainers](https://testcontainers.com/guides/getting-started-with-testcontainers-for-go/) provides a container setup process from within the go application.

Because Linux requires root (or special permissions), it may be necessary to run sudo to run tests. Because `sudo` may not have go install added to `PATH`, use `$(which go)` to pass the full path of the go command.
```shell
sudo $(which go) test -v
```
