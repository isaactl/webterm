package docker

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/isaactl/webterm/terminals"
	"github.com/isaactl/webterm/terminals/interface"
	"io"
	"log"
	"strings"
	"time"
)

type DockerClient struct {
	cli             *client.Client
	repo            string
	image           string
	containerID     string
	configs         terminals.TermConfigs
	SyncMessageFunc terminals.SyncFunc
	cmdBuff         bytes.Buffer
}

func NewDockerClient(configs terminals.TermConfigs) (_interface.Terminal, error) {
	repo := configs.Repo
	// set default repository
	if repo == "" {
		repo = "docker.io"
	}

	return &DockerClient{
		image:   configs.Image,
		repo:    repo,
		cmdBuff: bytes.Buffer{},
	}, nil
}

func (dc *DockerClient) SetSync(syncFunc terminals.SyncFunc) {
	dc.SyncMessageFunc = syncFunc
}

func (dc *DockerClient) Connect(context.Context) error {
	dc.SyncMessageFunc([]byte("Prepare environment...\r\n"), false)
	// should negotiate version with docker daemon
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		dc.SyncMessageFunc([]byte(err.Error()), true)
		return err
	}
	dc.cli = cli

	return dc.prepareContainer()
}

func (dc *DockerClient) Disconnect() error {
	dc.deleteContainer()
	if dc.cli != nil {
		dc.cli.Close()
	}
	return nil
}

func (dc *DockerClient) Run(cmd []byte) (int, error) {
	//fmt.Println(string(cmd))
	dc.SyncMessageFunc(cmd, false)
	if len(cmd) > 0 && cmd[0] != '\r' {
		dc.cmdBuff.Write(cmd)
		return 0, nil
	} else {
		dc.SyncMessageFunc([]byte("\n"), false)
		defer dc.cmdBuff.Reset()
	}

	cmdArray := strings.Split(dc.cmdBuff.String(), " ")
	fmt.Println(cmdArray)
	ctx := context.Background()
	exec, err := dc.cli.ContainerExecCreate(ctx, dc.containerID, types.ExecConfig{
		Privileged:   false,
		Tty:          true,
		AttachStdin:  true,
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          cmdArray,
	})
	if err != nil {
		return 0, err
	}

	/*	err = dc.cli.ContainerExecResize(ctx, exec.ID, types.ResizeOptions{
			Height: 100,
			Width:  100,
		})
		if err != nil {
			return nil, err
		}*/

	containerConn, err := dc.cli.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{
		Detach: false,
		Tty:    true,
	})
	if err != nil {
		return 0, err
	}

	go func() {
		defer func() {
			containerConn.Close()
		}()
		for {
			//docker reader and websocket writer
			buf := make([]byte, 4096)
			_, err = containerConn.Reader.Read(buf)
			if err != nil {
				log.Print(err)
				return
			}
			dc.SyncMessageFunc(buf, false)
			if err != nil {
				log.Print(err)
				dc.SyncMessageFunc([]byte(err.Error()), true)
				return
			}
		}
	}()

	return len(cmd), nil
}

func (dc *DockerClient) Read(msg []byte) (int, error) {
	r := "Hello World"
	copy(msg, []byte(r))
	time.Sleep(1 * time.Second)
	return len(r), nil
}

func (dc *DockerClient) Resize(resizeMessage terminals.WindowSize) error {
	return dc.cli.ContainerResize(context.Background(), dc.containerID, types.ResizeOptions{
		Height: uint(resizeMessage.Rows),
		Width:  uint(resizeMessage.Cols),
	})
}

func (dc *DockerClient) prepareContainer() error {
	ctx := context.Background()

	// check whether image exist
	images, err := dc.cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		dc.SyncMessageFunc([]byte(err.Error()), true)
		return err
	}

	foundMatch := false
	for _, image := range images {
		if strings.EqualFold(dc.image, image.ID) {
			foundMatch = true
			break
		}
		//fmt.Println(image.ID)
	}

	// TODO: add credential while pull image
	if !foundMatch {
		reader, err := dc.cli.ImagePull(ctx, fmt.Sprintf("%s/%s", dc.repo, dc.image), types.ImagePullOptions{})
		if err != nil {
			return err
		}
		var buff bytes.Buffer
		_, err = io.Copy(&buff, reader)
		if err != nil {
			dc.SyncMessageFunc([]byte(err.Error()), true)
			return err
		}
		dc.SyncMessageFunc(buff.Bytes(), false)
	}

	resp, err := dc.cli.ContainerCreate(ctx, &container.Config{
		Image: dc.image,
		Tty:   true,
	}, nil, nil, "")
	if err != nil {
		return err
	}

	if err := dc.cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	dc.containerID = resp.ID
	return nil
}

func (dc *DockerClient) deleteContainer() error {
	return nil
}
