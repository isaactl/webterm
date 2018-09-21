package _interface

import "github.com/isaactl/webterm/terminals"

type Terminal interface {
	Connect() error
	Disconnect() error
	Read([]byte) (int, error)
	Run(cmd []byte) error
	Resize(terminals.WindowSize) error
}
