package main

import (
    "os"

    "github.com/docker/docker/client"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/pkg/stdcopy"

    "golang.org/x/net/context"
    "log"
)

// should negotiate version with docker daemon
func NewDockerClient() (*client.Client, error) {
    return client.NewClientWithOpts(client.WithVersion("1.38"))
}

func Exec() {
    // ctx := context.Background()
}

func Demo() {
    ctx := context.Background()
    cli, err := NewDockerClient()
    if err != nil {
        panic(err)
    }

    resp, err := cli.ContainerCreate(ctx, &container.Config{
        AttachStderr: true,
        AttachStdin:  false,
        AttachStdout: true,
        //Tty:          true,
        Image: "ubuntu:16.04",
        Cmd:   []string{"/bin/bash"},
    }, nil, nil, "docker_client_test")
    if err != nil {
        panic(err)
    }

    if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
        panic(err)
    }

    statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
    select {
    case err := <-errCh:
        if err != nil {
            panic(err)
        }
    case <-statusCh:
    }

    out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
    if err != nil {
        panic(err)
    }

    stdcopy.StdCopy(os.Stdout, os.Stderr, out)

    exec, err := cli.ContainerExecCreate(ctx, resp.ID, types.ExecConfig{
        Privileged: false,
        Tty: false,
        AttachStdin: false,
        AttachStderr: false,
        AttachStdout: false,
        Detach: true,
        DetachKeys: "ctrl-p,ctrl-q",
        Cmd: []string{
            "date",
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    err = cli.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{
        Detach: true,
        Tty: false,
    })
    if err != nil {
        log.Fatal(err)
    }
}
