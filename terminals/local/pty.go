package local

import (
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
	SyncMessageFunc func([]byte)
	stopChan        chan bool
}

func NewPtyClinet(config terminals.TermConfigs, messageFunc func([]byte)) (_interface.Terminal, error) {
	return &PtyClient{
		SyncMessageFunc: messageFunc,
		stopChan:        make(chan bool, 1),
		//cmdBuff:         bytes.Buffer{},
	}, nil
}

func (client *PtyClient) Connect() error {
	client.SyncMessageFunc([]byte("lunch console\n"))
	client.cmd = exec.Command("/bin/bash", "-l")
	client.cmd.Env = append(os.Environ(), "TERM=xterm")

	tty, err := pty.Start(client.cmd)
	if err != nil {
		client.SyncMessageFunc([]byte(err.Error()))
		return err
	}
	client.tty = tty

	go func() {
		fmt.Println("Reading...")
		buf := make([]byte, 1024)
		for {
			select {
			case <-client.stopChan:
				return
			default:
				read, err := client.Read(buf)
				if err != nil {
					client.SyncMessageFunc([]byte(err.Error()))
					return
				}
				client.SyncMessageFunc(buf[:read])
			}
		}
	}()
	return nil
}

func (client *PtyClient) Disconnect() error {
	client.stopChan <- true
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

func (client *PtyClient) Run(cmd []byte) error {
	//fmt.Println(string(cmd))
	_, err := client.tty.Write(cmd)
	if err != nil {
		client.SyncMessageFunc([]byte(err.Error()))
	}
	return nil
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
		client.SyncMessageFunc([]byte("Unable to resize terminal"))
		return errors.New("Unable to resize terminal")
	}

	return nil
}
