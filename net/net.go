package net

import (
	"fmt"
	. "projet/message"
	"projet/utils"
	. "projet/utils"
)

const NB_SITES = 3

type Net struct {
	id       int
	clock    int
	tab      [NB_SITES]Request
	messages chan MessageWrapper
	server   Server
}

func NewNet(id int, port string, addr string) *Net {
	n := new(Net)
	n.id = id
	n.clock = 0
	var tab [NB_SITES]Request
	n.tab = tab
	n.server = *NewServer(port, addr, id, n)
  utils.Info(n.id, "NewNet", "Successfully created server instance")
	n.messages = make(chan MessageWrapper)
	return n
}

func (n Net) ReceiveCSrequest() {
  utils.Info(n.id, "ReceiveCSRequest", "received cs request from server")
	n.clock += 1
	n.tab[n.id] = Request{
		RequestType: "access",
		Stamp:       n.clock,
	}
  for i:=0; i < NB_SITES; i++ {
    utils.Error(i, "request state", n.tab[i].RequestType)
  }
	msg := Message{
		From:        n.id,
		To:          -1,
		Content:     "",
		Stamp:       n.clock,
		MessageType: "LockRequestMessage",
	}
	n.writeMessage(msg)
}

func (n Net) receiveCSRelease() {
	n.clock += 1
	n.tab[n.id] = Request{
		RequestType: "release",
		Stamp:       n.clock,
	}
  utils.Info(n.id, "MessageHandler", "Server CS release received")
	msg := Message{
		From:        n.id,
		To:          -1,
		Content:     "",
		Stamp:       n.clock,
		MessageType: "ReleaseMessage",
	}
	n.writeMessage(msg)
}

func (n Net) receiveExternalMessage(msg Message) {
	switch msg.MessageType {
	case "LockRequestMessage":
		n.receiveRequestMessage(msg)
	case "ReleaseMessage":
		n.receiveReleaseMessage(msg)
	case "AckMessage":
		n.receiveAckMessage(msg)
	}
}

func (n Net) receiveRequestMessage(received_msg Message) {
  utils.Info(n.id, "receiveRequestMessage", "Received request message")
	n.clock = Max(n.clock, received_msg.Stamp) + 1
	n.tab[received_msg.From] = Request{
		RequestType: "access",
		Stamp:       received_msg.Stamp,
	}
	msg := Message{
		From:        n.id,
		To:          received_msg.From,
		Content:     "",
		Stamp:       n.clock,
		MessageType: "AckMessage",
	}
  utils.Info(n.id, "receiveRequestMessage", "Sending Ack message")
	go n.writeMessage(msg)
	if n.tab[n.id].RequestType == "access" && n.isLastRequest() {
    n.server.SendMessage("OkCs")
    utils.Info(n.id, "MessageHandler", "Sending OkCs to server")
	}
}

func (n Net) isLastRequest() bool {
	for i := 0; i < NB_SITES; i++ {
		if i != n.id {
			if n.tab[i].Stamp < n.tab[n.id].Stamp {
        utils.Warning(n.id, "isLastRequest", "false")
				return false
			} else if n.tab[i].Stamp == n.tab[n.id].Stamp {
				if i < n.id {
          utils.Warning(n.id, "isLastRequest", "false")
					return false
				}
			}
		}
	}
  utils.Warning(n.id, "isLastRequest", "true")
	return true
}

func (n Net) receiveReleaseMessage(msg Message) {
  utils.Info(n.id, "receiveReleaseMessage", "Received release message")
	n.clock = Max(n.clock, msg.Stamp) + 1
	n.tab[msg.From] = Request{
		RequestType: "release",
		Stamp:       msg.Stamp,
	}
	if n.tab[n.id].RequestType == "access" && n.isLastRequest() {
    n.server.SendMessage("OkCs")
    utils.Info(n.id, "MessageHandler", "Sending OkCs to server")
	}
}

func (n Net) receiveAckMessage(msg Message) {
  utils.Info(n.id, "MessageHandler", "Received ack message")
	n.clock = Max(n.clock, msg.Stamp) + 1
	if n.tab[msg.From].RequestType == "release" {
		n.tab[msg.From] = Request{
			RequestType: "ack",
			Stamp:       msg.Stamp,
		}
	}
  utils.Warning(n.id, "receiveAckMessage", "request : " + n.tab[n.id].RequestType)
	if n.tab[n.id].RequestType == "access" && n.isLastRequest() {
    n.server.SendMessage("OkCs")
    utils.Info(n.id, "MessageHandler", "Sending OkCs to server")
	}
}

func (n Net) ReadMessage() {
	var raw string
	for {
		fmt.Scanln(&raw)
    utils.Info(n.id, "ReadMessage", "Detected new message")
		var msg = MessageFromString(raw)
		if msg.From != n.id {
			n.messages <- MessageWrapper{
				Action:  "process",
				Message: msg,
			}
		}
	}
}

func (n Net) writeMessage(msg Message) {
		// utils.Info(n.id, "WriteMessage", "Writing new message on the queue")
	n.messages <- (MessageWrapper{
		Action:  "send",
		Message: msg,
	})
}

func (n Net) MessageHandler() {
	for {
		wrapperItem := <-n.messages
		var msg = wrapperItem.Message
    utils.Info(n.id, "MessageHandler", "New message on the queue: " + msg.ToString())
		if wrapperItem.Action == "send" {
			utils.Info(n.id, "MessageHandler", "Spreading message on the ring")
			msg.Send()
		} else if wrapperItem.Action == "process" {
			if msg.To == n.id {
        utils.Info(n.id, "MessageHandler", "Handler processing the message")
				n.receiveExternalMessage(msg)
			} else if msg.To == -1 && msg.From != n.id { // message for all
				// process and spread the message on the ring
        utils.Info(n.id, "MessageHandler", "Handler processing and sending the message")
				n.receiveExternalMessage(msg)
				msg.Send()
			} else { // message not for us
				// forward the message
        utils.Info(n.id, "MessageHandler", "Handler forwarding the message")
				msg.Send()
			}
		}
    for i:=0; i < NB_SITES; i++ {
      utils.Error(i, "handler req", n.tab[i].RequestType)
    }
	}
}

// func (n Net) sendMessageToServer(msg Message) {
// 	utils.Info(n.id, "sendMessageToServer", msg.MessageType)
// 	n.server.SendMessage(msg.Content)
// }

func (n Net) sendMessageFromServer(msg Message) {
	utils.Info(n.id, "sendMessageFromServer", msg.MessageType)
	n.writeMessage(msg)
}
