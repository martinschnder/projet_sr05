package websocket

import (
	"fmt"
    "net/http"
	"github.com/gorilla/websocket"
	"appmodule/display"
	"sync"
)

var ws *websocket.Conn = nil
var Mutex = &sync.Mutex{}
var Message string = "hello"

func Do(w http.ResponseWriter, r *http.Request) {
    var upgrader = websocket.Upgrader{
        CheckOrigin : func(r *http.Request) bool { return true },
    }

    cnx,err := upgrader.Upgrade(w, r, nil)
    
	if err != nil {
        display.Error("ws_create", "Upgrade error : " + string(err.Error()))
        return
    }

    display.Info("ws_create", "Client connecté")

    ws = cnx
    go receive()  
}

func Send(msg string) {
    if ( ws == nil ) {
         fmt.Println("ws_send", "websocket non ouverte")
    } else {
        err := ws.WriteMessage( websocket.TextMessage,  []byte(msg) )
        if err != nil {
            fmt.Println("ws_send", "WriteMessage : " + string(err.Error()))
        } else {
            fmt.Println("ws_send", "sending " + msg)
        }
    }
}

func close() {
    display.Info("ws_close", "Fin des réceptions => fermeture de la websocket")
    ws.Close()
}

func receive() {
    defer close()

    for {
        _, rcvmsg, err := ws.ReadMessage()

        if err != nil {
            display.Error("ws_receive", "ReadMessage : " + string(err.Error()))
            break
        }

        Mutex.Lock()
      
        display.Info("ws_receive", "réception : " + string(rcvmsg[:]) )
        Message = string(rcvmsg)
               
        Mutex.Unlock()
    }
}