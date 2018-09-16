package main

import (
    "os"

    "github.com/docker/docker/client"
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "golang.org/x/net/context"
    "io"
)

// should negotiate version with docker daemon
func NewDockerClient() (*client.Client, error) {
    return client.NewClientWithOpts(client.WithVersion("1.38"))
}

func PrepareContainer() (string, error){
    ctx := context.Background()
    cli, err := NewDockerClient()
    if err != nil {
        return "", err
    }

    reader, err := cli.ImagePull(ctx, "docker.io/ubuntu:16.04", types.ImagePullOptions{})
    if err != nil {
        return "", err
    }
    io.Copy(os.Stdout, reader)

    resp, err := cli.ContainerCreate(ctx, &container.Config{
        Image: "ubuntu:16.04",
        Tty:true,
    }, nil, nil, "")
    if err != nil {
        return "", err
    }

    if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
        return "", err
    }

    return resp.ID, nil
}
