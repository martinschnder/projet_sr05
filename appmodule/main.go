package main

import (
	"appmodule/message"
	"appmodule/websocket"
	"appmodule/display"
	"time"
	"strconv"
	"flag"
	"net/http"
	"fmt"
)

var h int = 0;

func recaler(x, y int) int {
    if x < y {
        return y + 1
    }
    return x + 1
}

func sendperiodic() {
    for {
        websocket.Mutex.Lock()

        h = h + 1
        display.Display_d("sendperiodic", "h = " + strconv.Itoa(h))
        message.Msg_send(message.Msg_format("msg", websocket.Message) + message.Msg_format("hlg", strconv.Itoa(h)) )

        websocket.Ws_send(strconv.Itoa(h))
        
        websocket.Mutex.Unlock()
        time.Sleep(time.Duration(4) * time.Second)
    }
}

func receive() {
    var rcvmsg string
    var s_hrcv string
    var hrcv int

    for {
        fmt.Scanln(&rcvmsg)
        websocket.Mutex.Lock()

        display.Display_d("receive", "received msg is " + rcvmsg)

        s_hrcv = message.Findval(rcvmsg, "hlg")
        if s_hrcv != "" {
            hrcv, _ = strconv.Atoi(s_hrcv)
        }
        
        h = recaler(h, hrcv);
        display.Display_d("receive", "now h = " + strconv.Itoa(h))
        websocket.Ws_send(strconv.Itoa(h))
        
        websocket.Mutex.Unlock()
        rcvmsg = ""
    }
}

func main() {
    var port = flag.String("port", "4444", "n° de port")
    var addr = flag.String("addr", "localhost", "nom/adresse machine")
    var p_id = flag.String("id", "anonyme", "nom de l'instance")

    flag.Parse()
    display.Id = *p_id

    go sendperiodic()
    go receive()

    http.HandleFunc("/ws", websocket.Do_websocket)
    http.ListenAndServe(*addr + ":" + *port, nil)   
}