package local

import (
	"context"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/kr/pty"

	//"bytes"
	"fmt"
	"github.com/isaactl/webterm/terminals"
	"github.com/isaactl/webterm/terminals/interface"
	"github.com/pkg/errors"
)

type PtyClient struct {
	cmd *exec.Cmd
	tty *os.File
	// cmdBuff         bytes.Buffer
	SyncMessageFunc terminals.SyncFunc
}

func NewPtyClient(config terminals.TermConfigs) (_interface.Terminal, error) {
	return &PtyClient{
		//cmdBuff:         bytes.Buffer{},
	}, nil
}

func (client *PtyClient) SetSync(syncFunc terminals.SyncFunc) {
	client.SyncMessageFunc = syncFunc
}

func (client *PtyClient) Connect(ctx context.Context) error {
	client.SyncMessageFunc([]byte("lunch console\n"), false)
	client.cmd = exec.Command("/bin/bash", "-l")
	client.cmd.Env = append(os.Environ(), "TERM=xterm")

	tty, err := pty.Start(client.cmd)
	if err != nil {
		client.SyncMessageFunc([]byte(err.Error()), true)
		return err
	}
	client.tty = tty

	go func() {
		fmt.Println("Reading...")
		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				read, err := client.Read(buf)
				if err != nil {
					client.SyncMessageFunc([]byte(err.Error()), true)
					return
				}
				client.SyncMessageFunc(buf[:read], false)
			}
		}
	}()
	return nil
}

func (client *PtyClient) Disconnect() error {
	if client.cmd != nil {
		client.cmd.Process.Kill()
		client.cmd.Process.Wait()
	}

	if client.tty != nil {
		client.tty.Close()
	}
	return nil
}

func (client *PtyClient) Read(buf []byte) (int, error) {
	bytesRead, err := client.tty.Read(buf)
	if err != nil {
		return 0, err
	}

	return bytesRead, nil
}

func (client *PtyClient) Run(cmd []byte) (int, error) {
	//fmt.Println(string(cmd))
	_, err := client.tty.Write(cmd)
	if err != nil {
		client.SyncMessageFunc([]byte(err.Error()), true)
	}
	return len(cmd), nil
}

func (client *PtyClient) Resize(resizeMessage terminals.WindowSize) error {
	fmt.Printf("Resize %+v\n", resizeMessage)
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		client.tty.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&resizeMessage)),
	)
	if errno != 0 {
		client.SyncMessageFunc([]byte("Unable to resize terminal"), true)
		return errors.New("Unable to resize terminal")
	}

	return nil
}
