package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/isaactl/webterm/terminals"
	"github.com/isaactl/webterm/terminals/interface"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var Client _interface.Terminal

func handleWebsocket(w http.ResponseWriter, r *http.Request) {
	l := log.WithField("remoteaddr", r.RemoteAddr)
	if Client == nil {
		l.Error("client is not initialized")
		w.Write([]byte("client is not initialized"))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		l.WithError(err).Error("Unable to upgrade connection")
		return
	}
	defer func() {
		conn.Close()
	}()

	ctx, cancelFunc := context.WithCancel(context.Background())
	Client.SetSync(func(msg []byte, err bool) {
		conn.WriteMessage(websocket.BinaryMessage, msg)
		if err {
			conn.WriteMessage(websocket.CloseMessage, []byte("exiting..."))
			cancelFunc()
		}
	})
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("failed to connect to remote server"))
		l.WithError(err).Error("unable to create docker client")
		return
	}

	err = Client.Connect(ctx)
	if err != nil {
		l.WithError(err).Error("unable to connect server terminal")
		return
	} else {
		defer func() {
			Client.Disconnect()
		}()
	}

	handleMessage(ctx, conn)

	l.Println("stop")
}

func handleMessage(ctx context.Context, conn *websocket.Conn) {
	if ctx == nil || conn == nil {
		log.Fatal("arguments can't be nil")
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			messageType, reader, err := conn.NextReader()
			if err != nil {
				log.Error("Unable to grab next reader")
				return
			}

			if messageType == websocket.TextMessage {
				log.Warn("Unexpected text message")
				conn.WriteMessage(websocket.TextMessage, []byte("Unexpected text message"))
				continue
			}

			dataTypeBuf := make([]byte, 1)
			read, err := reader.Read(dataTypeBuf)
			if err != nil {
				log.Error("Unable to read message type from reader")
				conn.WriteMessage(websocket.TextMessage, []byte("Unable to read message type from reader"))
				return
			}

			if read != 1 {
				log.Error("Unexpected number of bytes read")
				return
			}

			switch dataTypeBuf[0] {
			case 0: // cmd data
				b, _ := ioutil.ReadAll(reader)
				fmt.Println(string(b))
				_, err := Client.Run(b)
				if err != nil {
					log.Errorf("Error after copying %v bytes", err)
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
				Client.Resize(resizeMessage)
			default:
				log.Error("Unknown data type")
			}
		}
	}
}

func lunchServer(listen, assetsPath string) {
	r := mux.NewRouter()

	r.HandleFunc("/term", handleWebsocket)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(assetsPath)))

	log.Info("Demo Websocket/Xterm terminal")
	log.Warn("Warning, this is a completely insecure daemon that permits anyone to connect and control your computer, please don't run this anywhere")

	if !(strings.HasPrefix(listen, "127.0.0.1") || strings.HasPrefix(listen, "localhost")) {
		log.Warn("Danger Will Robinson - This program has no security built in and should not be exposed beyond localhost, you've been warned")
	}

	if err := http.ListenAndServe(listen, r); err != nil {
		log.WithError(err).Fatal("Something went wrong with the webserver")
	}
}
