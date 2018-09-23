package vm

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/isaactl/webterm/terminals"
	"github.com/isaactl/webterm/terminals/interface"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type VMClient struct {
	client          *ssh.Client
	session         *ssh.Session
	sshWriter       io.Writer
	sshReader       io.Reader
	RemoteAdd       string
	Port            string
	UserName        string
	Password        string
	CredentialFile  string
	SyncMessageFunc terminals.SyncFunc
	stopChan        chan bool
	cmdBuff         bytes.Buffer
}

func NewVMClient(configs terminals.TermConfigs) (_interface.Terminal, error) {
	return &VMClient{
		RemoteAdd: configs.RemoteAdd,
		Port:      configs.Port,
		UserName:  configs.UserName,
		Password:  configs.Password,
		stopChan:  make(chan bool, 1),
	}, nil
}

func (vm *VMClient) SetSync(syncFunc terminals.SyncFunc) {
	vm.SyncMessageFunc = syncFunc
}

func (vm *VMClient) Connect(ctx context.Context) error {
	log.Printf("loging ...")
	vm.SyncMessageFunc([]byte("prepare env...\r\n"), false)
	// TODO: find a way to get hostkey instead of load it from file
	hostKey, err := loadHostKey(vm.RemoteAdd)
	if err != nil || hostKey == nil {
		return errors.New(fmt.Sprintf("hostkey not found: %v", err))
	}

	log.Printf("login with %s: %s", vm.UserName, vm.Password)
	config := &ssh.ClientConfig{
		User: vm.UserName,
		Auth: []ssh.AuthMethod{
			ssh.Password(vm.Password),
		},
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	}
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%s", vm.RemoteAdd, vm.Port), config)
	if err != nil || client == nil {
		vm.SyncMessageFunc([]byte("failed to connect to server"), true)
		return errors.New(fmt.Sprintf("failed to connect to %s: %v", vm.RemoteAdd, err))
	}
	vm.client = client

	session, err := client.NewSession()
	if err != nil || session == nil {
		vm.SyncMessageFunc([]byte("failed to start a new session"), true)
		return errors.New(fmt.Sprintf("failed to create new session %s: %v", vm.RemoteAdd, err))
	}
	vm.session = session

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		vm.SyncMessageFunc([]byte(err.Error()), true)
		return err
	}

	sshReader, err := session.StdoutPipe()
	if err != nil {
		vm.SyncMessageFunc([]byte(err.Error()), true)
		return err
	}
	sshWriter, err := session.StdinPipe()
	if err != nil {
		vm.SyncMessageFunc([]byte(err.Error()), true)
		return err
	}
	vm.sshWriter = sshWriter
	vm.sshReader = sshReader

	if err := session.Shell(); err != nil {
		vm.SyncMessageFunc([]byte(err.Error()), true)
		return err
	}

	// read from terminal
	go func() {
		fmt.Println("Reading...")
		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				read, err := vm.Read(buf)
				fmt.Println(string(buf[:read]))
				fmt.Print(buf[read])
				if err != nil {
					vm.SyncMessageFunc([]byte(err.Error()), true)
					return
				}
				vm.SyncMessageFunc(buf[:read], false)
			}
		}
	}()

	return nil
}

func (vm *VMClient) Disconnect() error {
	if vm.session != nil {
		vm.session.Close()
		vm.session.Wait()
	}

	if vm.client != nil {
		vm.client.Close()
		vm.client.Wait()
	}

	return nil
}

func (vm *VMClient) Read(buff []byte) (int, error) {
	if vm.sshReader != nil {
		return vm.sshReader.Read(buff)
	}

	return 0, errors.New("can't read from terminal")
}

func (vm *VMClient) Run(cmd []byte) (int, error) {
	if vm.sshWriter != nil {
		// TODO: handle special input "enter & tab"
		if cmd[0] == '\r' {
			vm.SyncMessageFunc(append(cmd, '\n'), false)
		} else {
			vm.SyncMessageFunc(cmd, false)
		}
		return vm.sshWriter.Write(cmd)
	}

	return 0, errors.New("can't write to terminal")
}

func (vm *VMClient) Resize(size terminals.WindowSize) error {
	vm.session.WindowChange(int(size.Rows), int(size.Cols))
	return nil
}

// https://github.com/golang/crypto/blob/master/ssh/example_test.go#L143
func loadHostKey(host string) (ssh.PublicKey, error) {
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], host) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				log.Fatalf("error parsing %q: %v", fields[2], err)
			}
			break
		}
	}

	if hostKey == nil {
		return nil, errors.New(fmt.Sprintf("no hostkey for %s", host))
	}

	return hostKey, nil
}
