package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/isaactl/webterm/terminal"
	"github.com/isaactl/webterm/terminal/docker"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

type windowSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
	X    uint16
	Y    uint16
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	l := log.WithField("remoteaddr", r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		l.WithError(err).Error("Unable to upgrade connection")
		return
	}

	configs := terminal.TermConfigs{
		RemoteAdd: "",
	}
	configs.Image = "ubuntu:16.04"

	client, err := docker.NewDockerClient(configs)
	if err != nil {
		l.WithError(err).Error("unable to create docker client")
		return
	}
	err = client.Connect()
	if err != nil {
		l.WithError(err).Error("unable to connect docker client")
		return
	}

	defer func() {
		client.Disconnect()
		conn.Close()
	}()
	client.Resize(uint(100), uint(100))

	go func() {
		for {
			buf := make([]byte, 1024)
			read, err := client.Read(buf)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
				l.WithError(err).Error("Unable to read from pty/cmd")
				return
			}
			if read > 0 {
				conn.WriteMessage(websocket.BinaryMessage, buf[:read])
			}
		}
	}()

	for {
		messageType, reader, err := conn.NextReader()
		if err != nil {
			l.WithError(err).Error("Unable to grab next reader")
			return
		}

		if messageType == websocket.TextMessage {
			l.Warn("Unexpected text message")
			conn.WriteMessage(websocket.TextMessage, []byte("Unexpected text message"))
			continue
		}

		dataTypeBuf := make([]byte, 1)
		read, err := reader.Read(dataTypeBuf)
		if err != nil {
			l.WithError(err).Error("Unable to read message type from reader")
			conn.WriteMessage(websocket.TextMessage, []byte("Unable to read message type from reader"))
			return
		}

		if read != 1 {
			l.WithField("bytes", read).Error("Unexpected number of bytes read")
			return
		}

		switch dataTypeBuf[0] {
		case 0: // cmd data
			b, _ := ioutil.ReadAll(reader)
			res, err := client.Run(conn, b)
			if err != nil {
				l.WithError(err).Errorf("Error after copying %v bytes", err)
			}
			conn.WriteMessage(websocket.BinaryMessage, res)
		case 1: // resize
			decoder := json.NewDecoder(reader)
			resizeMessage := windowSize{}
			err := decoder.Decode(&resizeMessage)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte("Error decoding resize message: "+err.Error()))
				continue
			}
			log.WithField("resizeMessage", resizeMessage).Info("Resizing terminal")
			client.Resize(uint(resizeMessage.Cols), uint(resizeMessage.Rows))
		default:
			l.WithField("dataType", dataTypeBuf[0]).Error("Unknown data type")
		}
	}
}

func main() {
	var listen = flag.String("listen", "127.0.0.1:3000", "Host:port to listen on")
	var assetsPath = flag.String("assets", "./view", "Path to assets")

	flag.Parse()

	r := mux.NewRouter()

	r.HandleFunc("/term", handleWebsocket)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(*assetsPath)))

	log.Info("Demo Websocket/Xterm terminal")
	log.Warn("Warning, this is a completely insecure daemon that permits anyone to connect and control your computer, please don't run this anywhere")

	if !(strings.HasPrefix(*listen, "127.0.0.1") || strings.HasPrefix(*listen, "localhost")) {
		log.Warn("Danger Will Robinson - This program has no security built in and should not be exposed beyond localhost, you've been warned")
	}

	if err := http.ListenAndServe(*listen, r); err != nil {
		log.WithError(err).Fatal("Something went wrong with the webserver")
	}
}
