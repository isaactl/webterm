package main

import (
	"encoding/json"
	"flag"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/isaactl/webterm/terminals"
	"github.com/isaactl/webterm/terminals/local"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

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
	defer func() {
		conn.Close()
	}()

	sendMsgToWebTerm := func(msg []byte) {
		//fmt.Println(string(msg))
		conn.WriteMessage(websocket.BinaryMessage, msg)
	}

	configs := terminals.TermConfigs{
		RemoteAdd: "",
	}
	// TODO: config base image
	configs.Image = "ubuntu:16.04"

	// client, err := docker.NewDockerClient(configs, sendMsgToWebTerm)
	client, err := local.NewPtyClinet(configs, sendMsgToWebTerm)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("failed to connect to remote server"))
		l.WithError(err).Error("unable to create docker client")
		return
	}

	err = client.Connect()
	if err != nil {
		l.WithError(err).Error("unable to connect docker client")
		return
	} else {
		defer func() {
			client.Disconnect()
		}()
	}

	// client.Resize(uint(100), uint(100))
	/*
		go func() {
			for {
				buf := make([]byte, 1024)
				read, err := client.Read(buf)
				if err != nil {
					fmt.Println(string(buf))
					conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
					l.WithError(err).Error("Unable to read from pty/cmd")
					return
				}
				if read > 0 {
					conn.WriteMessage(websocket.BinaryMessage, buf[:read])
				}
			}
		}()
	*/
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
			err := client.Run(b)
			if err != nil {
				l.WithError(err).Errorf("Error after copying %v bytes", err)
			}
			log.WithField("read from web term", string(b)).Infof("")
		case 1: // resize
			decoder := json.NewDecoder(reader)
			resizeMessage := terminals.WindowSize{}
			err := decoder.Decode(&resizeMessage)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte("Error decoding resize message: "+err.Error()))
				continue
			}
			log.WithField("resizeMessage", resizeMessage).Info("Resizing terminal")
			client.Resize(resizeMessage)
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
