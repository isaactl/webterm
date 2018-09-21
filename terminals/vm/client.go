package vm

import (
	"bytes"
	"github.com/isaactl/webterm/terminals"
	"github.com/isaactl/webterm/terminals/interface"
)

type VMClient struct {
	HostName        string
	UserName        string
	Password        string
	CredentialFile  string
	SyncMessageFunc func([]byte)
	cmdBuff         bytes.Buffer
}

func NewVMClient(configs terminals.TermConfigs, messageFunc func([]byte)) (_interface.Terminal, error) {
	return &VMClient{
		HostName:        configs.RemoteAdd,
		UserName:        configs.UserName,
		Password:        configs.Password,
		SyncMessageFunc: messageFunc,
	}, nil
}

func (vm *VMClient) Connect() error {
	return nil
}

func (vm *VMClient) Disconnect() error {
	return nil
}

func (vm *VMClient) Read([]byte) (int, error) {
	return 0, nil
}

func (vm *VMClient) Run(cmd []byte) error {
	return nil
}

func (vm *VMClient) Resize(size terminals.WindowSize) error {
	return nil
}
