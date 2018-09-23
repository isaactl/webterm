package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"fmt"
	"github.com/isaactl/webterm/terminals"
	"github.com/isaactl/webterm/terminals/docker"
	"github.com/isaactl/webterm/terminals/local"
	"github.com/isaactl/webterm/terminals/vm"
	"github.com/pkg/errors"
	"log"
	"os/exec"
	"runtime"
)

func main() {
	app := cli.NewApp()
	app.Name = "web termination "
	app.Usage = "./webterm"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "terminal",
			Usage: "run term in local/vm/container",
		},
		cli.StringFlag{
			Name:   "host",
			Usage:  "terminal host address",
			EnvVar: "HOST_ADD",
		},
		cli.StringFlag{
			Name:  "user",
			Usage: "user name for login",
		},
		cli.StringFlag{
			Name:  "password",
			Usage: "password for login",
		},
		cli.StringFlag{
			Name:  "repo",
			Usage: "repository for docker",
		},
		cli.StringFlag{
			Name:  "image",
			Usage: "docker image to start a container",
		},
		cli.StringFlag{
			Name:  "container-id",
			Usage: "existing container id",
		},
		cli.StringFlag{
			Name:  "port",
			Usage: "service listen port",
			Value: "3000",
		},
		cli.StringFlag{
			Name:  "ssl-port",
			Usage: "service listen port",
			Value: "22",
		},
		cli.StringFlag{
			Name:  "assets",
			Usage: "path to assets",
			Value: "./view",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	termConfigs := terminals.TermConfigs{
		RemoteAdd: c.String("host"),
		Port:      c.String("ssl-port"),
		UserName:  c.String("user"),
		Password:  c.String("password"),
	}

	termConfigs.Repo = c.String("repo")
	termConfigs.Image = c.String("image")
	termConfigs.ContainerID = c.String("container-id")

	var err error
	if terminal := c.String("terminal"); terminal != "" {
		switch terminal {
		case "localhost":
			Client, err = local.NewPtyClient(termConfigs)
		case "vm":
			Client, err = vm.NewVMClient(termConfigs)
		case "container":
			Client, err = docker.NewDockerClient(termConfigs)
		default:
			return errors.New("terminal should only be localhost/vm/container")
		}

		if err != nil {
			return err
		}
	} else {
		return errors.New("need to provide terminal")
	}

	listenAdd := fmt.Sprintf("localhost:%s", c.String("port"))
	log.Printf("server listen to %s", listenAdd)
	err = openBrowser(listenAdd)
	if err != nil {
		return err
	}

	lunchServer(listenAdd, c.String("assets"))

	return nil
}

func openBrowser(url string) error {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}

	return err
}
