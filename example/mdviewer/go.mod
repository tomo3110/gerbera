module github.com/tomo3110/gerbera/example/mdviewer

go 1.22

require (
	github.com/tomo3110/gerbera v0.0.0
	github.com/yuin/goldmark v1.7.8
)

require github.com/gorilla/websocket v1.5.3 // indirect

replace github.com/tomo3110/gerbera => ../..
