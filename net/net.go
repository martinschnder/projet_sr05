package net

import (
	"fmt"
	. "projet/message"
	"projet/utils"
	. "projet/utils"
	"encoding/json"
	"io/ioutil"
	// "time"
)

const NB_SITES = 3

type Net struct {
	id       int
	clock    int
	tab      [NB_SITES]Request
	messages chan MessageWrapper
	server   *Server
	color 	 string
	initator bool
	nbExpectedStates int
	nbExpectedMessages int
	state 	*State
	globalStates [NB_SITES]State
}

func NewNet(id int, port string, addr string) *Net {
	n := new(Net)
	n.id = id
	n.clock = 0
	var tab [NB_SITES]Request
	n.tab = tab
	for i := 0; i < NB_SITES; i++ {
		n.tab[i] = Request{
			RequestType: "release",
			Stamp:       0,
		}
	}
	n.server = NewServer(port, addr, id, n)
  	utils.Info(n.id, "NewNet", "Successfully created server instance")
	n.messages = make(chan MessageWrapper)
	n.color = "white"
	n.initator = false
	n.nbExpectedStates = 0
	n.nbExpectedMessages = 0
	n.state = NewState(id, n.server.data, NB_SITES)
	var gStates [NB_SITES]State
	n.globalStates = gStates
	return n
}

func (n *Net) ReceiveCSrequest() {
  utils.Info(n.id, "ReceiveCSRequest", "received cs request from server")
	n.clock += 1
	n.tab[n.id] = Request{
		RequestType: "access",
		Stamp:       n.clock,
	}

	msg := Message{
		From:        n.id,
		To:          -1,
		Content:     "",
		Stamp:       n.clock,
		MessageType: "LockRequestMessage",
		VectClock: 	 n.state.VectClock,
		Color: 		 n.color,
	}
	n.writeMessage(msg)
}

func (n *Net) receiveCSRelease() {
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
		VectClock: 	 n.state.VectClock,
		Color: 		 n.color,
	}
	go n.writeMessage(msg)
}

func (n *Net) receiveExternalMessage(msg Message) {
	n.state.VectClockIncr(msg.VectClock, NB_SITES)
	
	array := fmt.Sprint(n.state.VectClock)
	utils.Error(n.id, "receiveExternalMessage", array)

	switch msg.MessageType {
	case "LockRequestMessage":
		n.receiveRequestMessage(msg)
	case "ReleaseMessage":
		n.receiveReleaseMessage(msg)
	case "AckMessage":
		n.receiveAckMessage(msg)
  case "EditMessage":
    n.server.SendMessage(msg.Content)
case "SnapshotMessage":
    n.receiveSnapshotMessage(msg)
	}
}

func (n *Net) receiveRequestMessage(received_msg Message) {
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
		VectClock: 	 n.state.VectClock,
		Color: 		 n.color,
	}
  utils.Info(n.id, "receiveRequestMessage", "Sending Ack message")
	go n.writeMessage(msg)
	if n.tab[n.id].RequestType == "access" && n.isLastRequest() {
    n.server.SendMessage("OkCs")
    utils.Info(n.id, "MessageHandler", "Sending OkCs to server")
	}
}

func (n *Net) isLastRequest() bool {
	for i := 0; i < NB_SITES; i++ {
		if i != n.id {
      if n.tab[i].RequestType == "Request" {
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
	}
  utils.Warning(n.id, "isLastRequest", "true")
	return true
}

func (n *Net) receiveReleaseMessage(msg Message) {
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

func (n *Net) receiveAckMessage(msg Message) {
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
    utils.Info(n.id, "MessageHandler", "Sending OkCs to server")
    n.server.SendMessage("OkCs")
	}
}

func (n *Net) receiveSnapshotMessage(msg Message) {
	utils.Info(n.id, "receiveSnapshotMessage", "Received snapshot message")
	if (msg.Color == "red" && n.color == "white") {
		n.color = "red"
		n.globalStates[n.id] = *n.state
		msg := Message{
			From:        n.id,
			To:          -1,
			Content:     n.state.ToString(),
			Stamp:       n.clock,
			MessageType: "StateMessage",
			VectClock: 	 n.state.VectClock,
			Color: 		 n.color,
		}
		go n.writeMessage(msg)
	}
	if (msg.Color == "white" && n.color == "red") {
		msg.MessageType = "PrepostMessage"
		go n.writeMessage(msg)
	}
  }

  func (n *Net) receiveStateMessage(msg Message) {
	utils.Info(n.id, "receiveStateMessage", "Received state message")
	state := StateFromString(msg.Content)
	if (n.initator) {
		n.globalStates[state.Id] = state
		n.nbExpectedStates -= 1
		n.nbExpectedMessages += state.Review
		if (n.nbExpectedStates == 0 && n.nbExpectedMessages == 0) {
			file, _ := json.MarshalIndent(n.globalStates, "", " ")
			_ = ioutil.WriteFile("test.json", file, 0644)
		}
		n.initator = false
	} else {
		go n.writeMessage(msg)
	}
  }

  func (n *Net) receivePrepostMessage(msg Message) {
	utils.Info(n.id, "receivePrepostMessage", "Received prepost message")
	if (n.initator) {
		n.nbExpectedMessages -= 1
		utils.Error(n.id, "receivePrepostMessage", "concat prepost")
		if (n.nbExpectedStates == 0 && n.nbExpectedMessages == 0) {
			file, _ := json.MarshalIndent(n.globalStates, "", " ")
			_ = ioutil.WriteFile("test.json", file, 0644)
		}
		n.initator = false
	} else {
		go n.writeMessage(msg)
	}
  }

func (n *Net) ReadMessage() {
	var raw string
	for {
		fmt.Scanln(&raw)
    utils.Info(n.id, "ReadMessage", "Detected new message : " + raw)
		var msg = MessageFromString(raw)
		if msg.From != n.id {
			n.messages <- MessageWrapper{
				Action:  "process",
				Message: msg,
			}
		} else {
			n.state.Review -= 1
		}
	}
}

func (n *Net) writeMessage(msg Message) {
	utils.Info(n.id, "WriteMessage", "Writing new message on the queue")
	n.messages <- (MessageWrapper{
		Action:  "send",
		Message: msg,
	})
}

func (n *Net) MessageHandler() {
	for {
		wrapperItem := <-n.messages
		var msg = wrapperItem.Message
    	utils.Info(n.id, "MessageHandler", "New message on the queue: " + msg.ToString())
		if wrapperItem.Action == "send" {
			utils.Info(n.id, "MessageHandler", "Spreading message on the ring")
			n.state.Review += 1
			msg.Send()
		} else if wrapperItem.Action == "process" {
			if msg.To == n.id {
        		utils.Info(n.id, "MessageHandler", "Handler processing the message")
				n.state.Review -= 1
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
		// time.Sleep(time.Duration(1) * time.Second)
	}
}

// func (n Net) sendMessageToServer(msg Message) {
// 	utils.Info(n.id, "sendMessageToServer", msg.MessageType)
// 	n.server.SendMessage(msg.Content)
// }

func (n *Net) SendMessageFromServer(msg Message) {
	utils.Info(n.id, "SendMessageFromServer", msg.MessageType)
	go n.writeMessage(msg)
}

func (n *Net) InitSnapshot() {
	n.color = "red"
	n.initator = true
	n.globalStates[n.id] = *n.state
	n.nbExpectedStates = NB_SITES - 1
	n.nbExpectedMessages = n.state.Review
	utils.Info(n.id, "InitSnapshot", "Snapshot initialisation")

	msg := Message{
		From:        n.id,
		To:          -1,
		Content:     "",
		Stamp:       n.clock,
		MessageType: "SnapshotMessage",
		VectClock: 	 n.state.VectClock,
		Color: 		 n.color,
	}
	go n.writeMessage(msg)
}