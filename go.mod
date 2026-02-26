module github.com/aK1r4z/emi-transport

go 1.25.4

retract (
	v1.0.2 // 早期错误发布的版本，请使用 v0.0.1
	v1.0.1 // 早期错误发布的版本，请使用 v0.0.1
	v1.0.0 // 早期错误发布的版本，请使用 v0.0.1
)

require (
	github.com/aK1r4z/emi-core v0.0.1
	github.com/gorilla/websocket v1.5.3
)
