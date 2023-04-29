package net

import (
	"net/http"
	. "projet/message"
	"projet/utils"
	"sync"
	"github.com/gorilla/websocket"
)

type Server struct {
  socket *websocket.Conn
  mutex *sync.Mutex
  data []string
  id int
  net *Net
  command Command
}

func NewServer(port string, addr string, id int, net *Net) *Server {
	server := new(Server)
	// s.clock = 0
  server.socket = nil
  server.net = net
  server.mutex = &sync.Mutex{}
  server.data = make([]string, 30)
  server.data[0] = "Hello, World!"
  server.id = id

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
    server.Send()
}

func (server Server)Send() {
    if (server.socket == nil ) {
         utils.Error(server.id, "ws_send", "websocket non ouverte")
    } else {
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
        for i:=0; i < NB_SITES; i++ {
          utils.Error(i, "request state", server.net.tab[i].RequestType)
        }
        server.mutex.Lock()
        command := ParseCommand(string(rcvmsg))
        server.command = command
        utils.Info(server.id, "ws_receive", "réception : " + string(rcvmsg[:]) )
        server.net.ReceiveCSrequest()
        utils.Info(server.id, "ws_receive", "commande de type : " + string(command.Action) )
        server.mutex.Unlock()
    }
}

func (server Server)EditData(command Command) {
  switch command.Action {
  case "Replace":
    server.data[command.Line] = command.Content
  case "Add":
    server.data[command.Line] += command.Content
  case "Remove":
    server.data[command.Line] = ""
}
  server.forwardEdition(command)
  server.Send()
}

func (server Server)forwardEdition(command Command) {
  server.net.sendMessageFromServer(Message{
    From: server.net.id,
    To: -1,
    Content: command.ToString(),
    Stamp: server.net.clock,
    MessageType: "EditMessage",
  })
}

// Used by net to send a message to server
func (server Server)SendMessage(action string) {
  if action == "OkCs" {
    // Request accepted
    server.EditData(server.command)
  } else {
    // Incoming command from another site
    command := ParseCommand(action)
    server.EditData(command)
  }
}
