package websocket

import (
	"fmt"
    "net/http"
	"github.com/gorilla/websocket"
	"appmodule/display"
    "appmodule/message"
	"sync"
)

var ws *websocket.Conn = nil
var Mutex = &sync.Mutex{}
var Message string = `hello moi

c'est bobby ca va`

type messageToSend struct {
    Message string
    Horloge int
}

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

func Send(hlg int) {
    if ( ws == nil ) {
         fmt.Println("ws_send", "websocket non ouverte")
    } else {
        formattedMessage := messageToSend{
            Message,
            hlg,
        }
        err := ws.WriteJSON(&formattedMessage)
        if err != nil {
            fmt.Println("ws_send", "WriteMessage : " + string(err.Error()))
        } else {
            fmt.Printf("ws_send sending : %+v\n", formattedMessage)
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
        // message recu du type /=line=13/=action=Remplacer/=message=kugbj
        
        Mutex.Unlock()
    }
}