package docker

import (
	"os"

	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gorilla/websocket"
	"github.com/isaactl/webterm/terminal"
	"github.com/isaactl/webterm/terminal/interface"
	"io"
	"log"
	"strings"
	"time"
)

type DockerClient struct {
	cli         *client.Client
	repo        string
	image       string
	containerID string
	configs     terminal.TermConfigs
}

func NewDockerClient(configs terminal.TermConfigs) (_interface.Terminal, error) {
	repo := configs.Repo
	// set default repository
	if repo == "" {
		repo = "docker.io"
	}

	return &DockerClient{
		image: configs.Image,
		repo:  repo,
	}, nil
}

func (dc *DockerClient) Connect() error {
	// should negotiate version with docker daemon
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
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

func (dc *DockerClient) Run(conn *websocket.Conn, cmd []byte) ([]byte, error) {
	var cmdArray []string
	err := json.Unmarshal(cmd, &cmdArray)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exec, err := dc.cli.ContainerExecCreate(ctx, dc.containerID, types.ExecConfig{
		Privileged:   false,
		Tty:          false,
		AttachStdin:  false,
		AttachStderr: false,
		AttachStdout: false,
		Detach:       true,
		DetachKeys:   "ctrl-p,ctrl-q",
		Cmd:          cmdArray,
	})
	if err != nil {
		return nil, err
	}

	containerConn, err := dc.cli.ContainerExecAttach(ctx, exec.ID, types.ExecStartCheck{
		Detach: true,
		Tty:    false,
	})
	if err != nil {
		return nil, err
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
			err = conn.WriteMessage(websocket.BinaryMessage, buf)
			if err != nil {
				log.Print(err)
				conn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
				return
			}
		}
	}()

	return nil, nil
}

func (dc *DockerClient) Read(msg []byte) (int, error) {
	r := "Hello World"
	copy(msg, []byte(r))
	time.Sleep(1 * time.Second)
	return len(r), nil
}

func (dc *DockerClient) prepareContainer() error {
	ctx := context.Background()

	// check whether image exist
	images, err := dc.cli.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	foundMatch := false
	for _, image := range images {
		if strings.EqualFold(dc.image, image.ID) {
			foundMatch = true
		}
		fmt.Println(image.ID)
	}

	if !foundMatch {
		reader, err := dc.cli.ImagePull(ctx, fmt.Sprintf("%s/%s", dc.repo, dc.image), types.ImagePullOptions{})
		if err != nil {
			return err
		}
		io.Copy(os.Stdout, reader)
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
