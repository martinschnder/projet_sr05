package net

import (
	"net/http"
	. "projet/message"
	"projet/utils"
	"sync"
	"github.com/gorilla/websocket"
  "fmt"
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

func (server *Server)createSocket(w http.ResponseWriter, r *http.Request) {
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

func (server *Server)Send() {
    if (server.socket == nil ) {
         utils.Error(server.id, "ws_send", "websocket non ouverte")
    } else {
        err := server.socket.WriteJSON(server.data)
        if err != nil {
            utils.Error(server.id, "ws_send", "Error message : " + string(err.Error()))
        } else {
            utils.Info(server.id, "ws_send", "sending data to client")
        }
    }
}

func (server *Server)closeSocket() {
    utils.Info(server.id, "ws_close", "Fin des réceptions => fermeture de la websocket")
    server.socket.Close()
}

func (server *Server)receive() {
    defer server.closeSocket()
    for {
        _, rcvmsg, err := server.socket.ReadMessage()
        if err != nil {
            utils.Error(server.id, "ws_receive", "ReadMessage : " + string(err.Error()))
            break
        }
        
        server.mutex.Lock()
        command := ParseCommand(string(rcvmsg))
        if (command.Action == "Snapshot") {
          server.net.InitSnapshot()
        } else {
          server.command = command
          server.net.ReceiveCSrequest()
        }
        utils.Info(server.id, "ws_receive", "réception : " + string(rcvmsg[:]) )
        utils.Info(server.id, "ws_receive", "commande de type : " + string(command.Action) )
        server.mutex.Unlock()
    }
}

func (server *Server)EditData(command Command) {
  utils.Error(server.id, "EditData", "enter fct" )
  utils.Error(server.id, "Command", command.ToString())
  switch command.Action {
  case "Remplacer":
    server.data[command.Line - 1] = command.Content
  case "Ajouter":
    server.data[command.Line - 1] += command.Content
  case "Supprimer":
    server.data[command.Line - 1] = ""
}


  array := fmt.Sprint(server.data[0])
  utils.Error(server.id, "EditData", array)

  server.net.state.Text = server.data

  // server.forwardEdition(command)
  // server.net.receiveCSRelease()
  server.Send()
}

func (server *Server)forwardEdition(command Command) {
  utils.Error(server.id, "forwardEdition", "enter fct" )

  server.net.SendMessageFromServer(Message{
    From: server.net.id,
    To: -1,
    Content: command.ToString(),
    Stamp: server.net.clock,
    MessageType: "EditMessage",
    VectClock: 	 server.net.state.VectClock,
  })

  utils.Error(server.id, "forwardEdition", "fini fct" )

}

// Used by net to send a message to server
func (server *Server)SendMessage(action string) {
  if action == "OkCs" {
    utils.Info(server.id, "ServSendMessage", "OkCs received" )
    // Request accepted
    server.EditData(server.command)
    server.forwardEdition(server.command)
    server.net.receiveCSRelease()
  } else {
    // Incoming command from another site
    command := ParseCommand(action)
    server.EditData(command)
  }
}
