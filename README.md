# postgres
go wrappers, migrations, and testcontainers for postgres

## 🏗️ testcontainers
[Testcontainers](https://testcontainers.com/guides/getting-started-with-testcontainers-for-go/) provides a container setup process from within the go application.

Because Linux requires root (or special permissions) for the Docker socket, it may be necessary to run sudo to run tests. Because `sudo` may not have go install added to `PATH`, use `$(which go)` to pass the full path of the go command. This is not necessary for CI tests, which have preconfigured permissions to the Docker socket.
```shell
sudo $(which go) test -v
```

## ⚙️ migrations
Files for initailizing (and tearing down) a database are included in the `Migrations()` function. See [`example-migrations`](https://github.com/ddbgio/postgres/tree/main/example-migrations) for examples.

> [!TIP]
> Migrations files should be numbered in the order they are to be run, first `down`, then `up`. As such, all `down` migrations should be odd; all `up` migrations should be even. Migrations will return a list of `Migration` objects, containing _Direction_, _Filename_, and _Content_.

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
// first argument should be the path to directory with the migrations files,
// second argument should be the direction of the migraiton
migrations, err = Migrations("example-migrations", "down")
if err != nil {
    return nil, fmt.Errorf("unable to fetch migrations: %w", err)
}

// executing migrations is outside the scope of this package,
// but may be implemented as such
for _, migration := range migrations {
    slog.Info("running a migration",
        "direction", migration.Direction,
        "filename", migration.Filename,
    )
    err := db.Execute(migration.Content, nil)
    if err != nil {
        return nil, fmt.Errorf("unable to execute migration: %w", err)
    }
}
```
