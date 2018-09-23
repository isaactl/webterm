package _interface

import (
	"context"
	"github.com/isaactl/webterm/terminals"
)

type Terminal interface {
	Connect(ctx context.Context) error
	Disconnect() error
	Read([]byte) (int, error)
	Run(cmd []byte) (int, error)
	Resize(terminals.WindowSize) error
	SetSync(syncFunc terminals.SyncFunc)
}
