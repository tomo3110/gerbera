module github.com/tomo3110/gerbera/example/sns

go 1.22

replace github.com/tomo3110/gerbera => ../..

require (
	github.com/go-sql-driver/mysql v1.9.3
	github.com/tomo3110/gerbera v0.0.0-00010101000000-000000000000
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
)
