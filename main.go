package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/gorilla/websocket"
    "flag"
    "github.com/jasonsoft/napnap"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func parseCmdLine(cmd string) []string {
    //reduce multi spaces into one space
    regexMultiSpace := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
    cmd = strings.TrimSpace(cmd)
    cmd = regexMultiSpace.ReplaceAllString(cmd, " ")
    //split string to array by space
    cmdArray := strings.Split(cmd, " ")
    return cmdArray
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {

	//upgrade http to websocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print(err)
		return
	}
	defer conn.Close()

	cli, err := NewDockerClient()
	if err != nil {
		log.Print(err)
		conn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
		return
	}

	//get cmd variable
	cmdArray := parseCmdLine(r.FormValue("cmd"))
	ctx := context.Background()
	execConfig := types.ExecConfig{
		AttachStderr: true,
		AttachStdin:  true,
		AttachStdout: true,
		Cmd:          cmdArray,
		Tty:          true,
		Detach:       false,
	}

	//set target container
	exec, err := cli.ContainerExecCreate(ctx, "2cd6333dd391", execConfig)
	if err != nil {
		log.Print(err)
		conn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
		return
	}
	execAttachConfig := types.ExecStartCheck{
		Detach: false,
		Tty:    true,
	}
	containerConn, err := cli.ContainerExecAttach(ctx, exec.ID, execAttachConfig)
	if err != nil {
		log.Print(err)
		conn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
		return
	}
	defer containerConn.Close()

	go func() {
		defer func() {
			containerConn.Close()
			conn.Close()
		}()
		for {
			//docker reader and websocket writer
			buf := make([]byte, 4096)
			_, err = containerConn.Reader.Read(buf)
			if err != nil {
				log.Print(err)
				conn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
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

	for {
		//docker writer and websocket reader
		_, reader, err := conn.NextReader()
		if err != nil {
			log.Print(err)
			conn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
			return
		}
		n, err := io.Copy(containerConn.Conn, reader)
		println(n)
		if err != nil {
			log.Print(err)
			conn.WriteMessage(websocket.BinaryMessage, []byte(err.Error()))
			return
		}
	}
}

func launchTerm() {
    var listen = flag.String("listen", ":8000", "Host:port to listen on")
    nap := napnap.New()
    flag.Parse()
    router := napnap.NewRouter()
    router.Get("/term", napnap.WrapHandler(http.HandlerFunc(handleWebsocket)))
    nap.Use(router)
    httpengine := napnap.NewHttpEngine(*listen)
    log.Fatal(nap.Run(httpengine))
}

func main() {
	//Demo()
	launchTerm()
}
