package websocket

import (
	"fmt"
    "net/http"
	"github.com/gorilla/websocket"
	"time"
	"appmodule/display"
	"sync"
)

var ws *websocket.Conn = nil
var Mutex = &sync.Mutex{}
var Message string = "hello"

func Do_websocket(w http.ResponseWriter, r *http.Request) {
    var upgrader = websocket.Upgrader{
        CheckOrigin : func(r *http.Request) bool { return true },
    }

    cnx,err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        display.Display_e("ws_create", "Upgrade error : " + string(err.Error()))
        return
    }

    display.Display_d("ws_create", "Client connecté")

    ws = cnx
    go Ws_receive()  
}

func Do_webserver(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Bonjour depuis le serveur web en Go !")
}

func Do_send() {
	for {
		Ws_send("Hello depuis le serveur");
		time.Sleep(time.Duration(4) * time.Second)
	}
}

func Ws_send(msg string) {
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

func Ws_close() {
    display.Display_d("ws_close", "Fin des réceptions => fermeture de la websocket")
    ws.Close()
}

func Ws_receive() {
    defer Ws_close()

    for {
        _, rcvmsg, err := ws.ReadMessage()

        if err != nil {
            display.Display_e("ws_receive", "ReadMessage : " + string(err.Error()))
            break
        }

        Mutex.Lock()
      
        display.Display_d("ws_receive", "réception : " + string(rcvmsg[:]) )
        Message = string(rcvmsg)
               
        Mutex.Unlock()
    }
}