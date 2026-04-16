The most production-ready go folder structure

\*Some notes
air init - to generate .air.toml file
go mod init rest-api-in-gin - initializes module
go get <package_name>
go install <github.com/golang-migrate/migrate/v4/cmd/migrate@latest>
migrate create -ext sql -dir ./cmd/migrate/migrate/migrations -seq <create_users_table>
go run ./cmd/migrate/main.go up
(watch) air
(manual) go run cmd/main.go
