package net

import (
	"github.com/gorilla/websocket"
	"net/http"
	. "projet/message"
	"projet/utils"
)

type Server struct {
	socket  *websocket.Conn
	data    []string
	id      int
	net     *Net
	command Command
}

func NewServer(port string, addr string, id int, net *Net) *Server {
	server := new(Server)
	server.socket = nil
	server.net = net
	server.data = make([]string, 30)
	server.data[0] = "Hello World!"
	server.id = id

	go http.HandleFunc("/ws", server.createSocket)
	go http.ListenAndServe(addr+":"+port, nil)
	utils.Info(server.id, "newServer", "Server listening on "+addr+":"+port)
	return server
}

/** createSocket()
Crée le socket auquel se connectera le client web
**/
func (server *Server) createSocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	cnx, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.Error(server.id, "ws_create", "Upgrade error : "+string(err.Error()))
		return
	}
	utils.Info(server.id, "ws_create", "Client connecté")
	server.socket = cnx
	go server.receive()
	server.Send()
}

/** Send()
Envoie le message au client Web
**/
func (server *Server) Send() {
	if server.socket == nil {
		utils.Error(server.id, "ws_send", "Websocket is closed")
	} else {
    msg := MessageToClient{
      Data: server.data,
      Stamp: server.net.clock,
    }
		err := server.socket.WriteJSON(msg)
		if err != nil {
			utils.Error(server.id, "ws_send", "Error message : "+string(err.Error()))
		} else {
			utils.Info(server.id, "ws_send", "Sending data to client")
		}
	}
}

/** closeSocket()
Ferme le socket
**/
func (server *Server) closeSocket() {
  utils.Info(server.id, "ws_close", "End of receptions : closing socket")
	server.socket.Close()
}

/** receive()
Traite la réception d'un message en provenance du client Web
**/
func (server *Server) receive() {
	defer server.closeSocket()
	for {
		_, rcvmsg, err := server.socket.ReadMessage()
		if err != nil {
			utils.Error(server.id, "ws_receive", "Error : "+string(err.Error()))
			break
		}

		command := ParseCommand(string(rcvmsg))
		if command.Action == "Snapshot" {
			server.net.InitSnapshot()
		} else {
			// La commande est sauvegardée pour plus tard, une fois les acquittements reçus
			server.command = command
			server.net.ReceiveCSrequest()
		}
		utils.Info(server.id, "ws_receive", "Received "+string(command.Action) + " action")
	}
}

/** EditData()
Gère l'action et effectue la commande
**/
func (server *Server) EditData(command Command) {
	switch command.Action {
	case "Remplacer":
		server.data[command.Line-1] = command.Content
	case "Ajouter":
		server.data[command.Line-1] += command.Content
	case "Supprimer":
		server.data[command.Line-1] = ""
  default:
    utils.Error(server.id, "EditData", "Unknown command action")
	}
	server.net.state.Text = server.data
	server.Send()
}

/** forwardEdition()
Propage la modification auprès de tous les sites
**/
func (server *Server)forwardEdition(command Command) {
	utils.Info(server.id, "forwardEdition", "Server forwarding edition")
  server.net.SendMessageFromServer(Message{
    From: server.net.id,
    To: -1,
    Content: command.ToString(),
    Stamp: server.net.clock,
    MessageType: "EditMessage",
    VectClock: 	 server.net.state.VectClock,
    Color: server.net.color,
  }}

/** SendMessage()
Utilisée par le site. Selon le message, autorise la modification ou effectue une modification en provenance d'un autre site
**/
func (server *Server) SendMessage(action string) {
	if action == "OkCs" {
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
