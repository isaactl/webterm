package _interface

import "github.com/gorilla/websocket"

type Terminal interface {
	Connect() error
	Disconnect() error
	Read([]byte) (int, error)
	Run(conn *websocket.Conn, cmd []byte) ([]byte, error)
}
