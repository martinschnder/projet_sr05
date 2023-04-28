package net

import (
	"net/http"
	"projet/message"
	"projet/utils"

	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
  socket *websocket.Conn
  mutex *sync.Mutex
  data []string
  id int
  net *Net
}

// func (s Server) recaler(distant_hlg int) {
// 	if s.clock < distant_hlg {
// 		s.clock = distant_hlg + 1
// 	} else {
// 		s.clock += 1
// 	}
// }

func (server Server) sendPeriodic() {
	for {
		server.mutex.Lock()

		// server.clock = server.clock + 1
		utils.Info(server.id, "sendperiodic", "h = idk")
		// message.Send(message.Format("msg", websocket.Message) + message.Format("hlg", strconv.Itoa(s.clock)))
    utils.Info(server.id,"msg_send", "émission hello" )

		server.Send()

		server.mutex.Unlock()
		time.Sleep(time.Duration(4) * time.Second)
	}
}

// func (b Server) receive() {
// 	var rcvmsg string
// 	var s_hrcv string
// 	var hrcv int
//
// 	for {
// 		fmt.Scanln(&rcvmsg)
// 		websocket.Mutex.Lock()
//
// 		display.Info("receive", "received msg is "+rcvmsg)
//
// 		s_hrcv = message.Findval(rcvmsg, "hlg")
// 		if s_hrcv != "" {
// 			hrcv, _ = strconv.Atoi(s_hrcv)
// 		}
//
// 		s.recaler(hrcv)
// 		display.Info("receive", "now h = "+strconv.Itoa(s.clock))
// 		websocket.Send(s.clock)
//
// 		websocket.Mutex.Unlock()
// 		rcvmsg = ""
// 	}
// }
// func (b Server)AskCs() {
//   fmt.Println("Requested cs from bas")
//   s.net.ReceiveCSrequest()
// }

func NewServer(port string, addr string, id int, net *Net) *Server {
	server := new(Server)
	// s.clock = 0
  server.socket = nil
  server.net = net
  server.mutex = &sync.Mutex{}
  server.data = make([]string, 30)
  server.data[0] = "Hello, World!"
  server.id = id

	// go server.sendPeriodic()

	go http.HandleFunc("/ws", server.createSocket)
	go http.ListenAndServe(addr+":"+port, nil)
  utils.Info(server.id, "newServer", "server listening on " + addr + ":" + port)
  return server
}

func (server Server)createSocket(w http.ResponseWriter, r *http.Request) {
    var upgrader = websocket.Upgrader{
        CheckOrigin : func(r *http.Request) bool { return true },
    }
    cnx,err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        utils.Error(server.id, "ws_create", "Upgrade error : " + string(err.Error()))
        return
    }
    utils.Info(server.id, "ws_create", "Client connecté")
    server.socket = cnx
    go server.receive()
}

func (server Server)Send() {
    if (server.socket == nil ) {
         utils.Error(server.id, "ws_send", "websocket non ouverte")
    } else {
        // formattedMessage := messageToSend{
        //     socket.message,
        //     hlg,
        // }
        err := server.socket.WriteJSON(&server.data)
        if err != nil {
            utils.Error(server.id, "ws_send", "Error message : " + string(err.Error()))
        } else {
            utils.Info(server.id, "ws_send", "sending data to client")
        }
    }
}

func (server Server)closeSocket() {
    utils.Info(server.id, "ws_close", "Fin des réceptions => fermeture de la websocket")
    server.socket.Close()
}

func (server Server)receive() {
    defer server.closeSocket()
    for {
        _, rcvmsg, err := server.socket.ReadMessage()
        if err != nil {
            utils.Error(server.id, "ws_receive", "ReadMessage : " + string(err.Error()))
            break
        }
        server.mutex.Lock()
    command := message.ParseCommand(string(rcvmsg))
        utils.Info(server.id, "ws_receive", "réception : " + string(rcvmsg[:]) )

        // TODO handle message
        server.net.ReceiveCSrequest()
        utils.Info(server.id, "ws_receive", "commande de type : " + string(command.Action) )
        
        // server.Send()
        server.mutex.Unlock()
    }
}
