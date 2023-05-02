package net

import (
	"github.com/gorilla/websocket"
	"net/http"
	. "projet/utils"
	"projet/display"
)

type Server struct {
	socket  *websocket.Conn
	text    []string
	id      int
	net     *Net
	command Command
}

func NewServer(port string, addr string, id int, net *Net) *Server {
	server := new(Server)
	server.socket = nil
	server.net = net
	server.text = make([]string, 30)
	server.text[0] = "Hello World!"
	server.id = id

	go http.HandleFunc("/ws", server.createSocket)
	go http.ListenAndServe(addr+":"+port, nil)
	display.Info(server.id, "newServer", "Server listening on "+addr+":"+port)
	return server
}

/** createSocket()
Crée la websocket auquel se connectera le client web
**/
func (server *Server) createSocket(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	cnx, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		display.Error(server.id, "ws_create", "Upgrade error : "+string(err.Error()))
		return
	}
	display.Info(server.id, "ws_create", "Client connecté")
	server.socket = cnx
	go server.receive()
	server.Send()
}

/** Send()
Envoie le message au client Web
**/
func (server *Server) Send() {
	if server.socket == nil {
		display.Error(server.id, "ws_send", "Websocket is closed")
	} else {
    msg := MessageToClient{
      Text: server.text,
      Stamp: server.net.clock,
    }
		err := server.socket.WriteJSON(msg)
		if err != nil {
			display.Error(server.id, "ws_send", "Error message : "+string(err.Error()))
		} else {
			display.Info(server.id, "ws_send", "Sending data to client")
		}
	}
}

/** closeSocket()
Ferme la websocket
**/
func (server *Server) closeSocket() {
  display.Info(server.id, "ws_close", "End of receptions : closing socket")
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
			display.Error(server.id, "ws_receive", "Error : "+string(err.Error()))
			break
		}

		command := CommandFromString(string(rcvmsg))
		if command.Action == "Snapshot" {
			server.net.InitSnapshot()
		} else {
			// La commande est sauvegardée pour plus tard, une fois les acquittements reçus
			server.command = command
			server.net.ReceiveCSRequest()
		}
    display.Info(server.id, "ws_receive", "Received action :" + string(command.Action))
	}
}

/** EditText()
Gère l'action et effectue la commande
**/
func (server *Server) EditText(command Command) {
	switch command.Action {
	case "Remplacer":
		server.text[command.Line-1] = command.Content
	case "Ajouter":
		server.text[command.Line-1] += command.Content
	case "Supprimer":
		server.text[command.Line-1] = ""
  default:
    display.Error(server.id, "EditText", "Unknown command action")
    return
	}
	server.net.state.Text = server.text
	server.Send()
}

/** forwardEdition()
Propage la modification auprès de tous les sites
**/
func (server *Server)forwardEdition(command Command) {
	display.Info(server.id, "forwardEdition", "Server forwarding edition")
  server.net.SendMessageFromServer(Message{
    From: server.net.id,
    To: -1,
    Content: command.ToString(),
    Stamp: server.net.clock,
    MessageType: "EditMessage",
    VectClock: 	 server.net.state.VectClock,
    Color: server.net.color,
  })
}

/** SendMessage()
Utilisée par le site. Selon le message, autorise la modification ou effectue une modification en provenance d'un autre site
**/
func (server *Server) SendMessage(action string) {
	if action == "OkCs" {
		// Request accepted
		server.EditText(server.command)
		server.forwardEdition(server.command)
		server.net.receiveCSRelease()
		server.Send()
	} else {
		// Incoming command from another site
		command := CommandFromString(action)
		server.EditText(command)
	}
}
