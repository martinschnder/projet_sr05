package net

import (
	"fmt"
	. "projet/message"
	"projet/utils"
	. "projet/utils"
	"container/list"
	"os"
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
	globalStates *list.List
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
	n.globalStates = list.New()
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
	case "StateMessage":
    n.receiveStateMessage(msg)
	case "PrepostMessage":
    n.receivePrepostMessage(msg)
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
	utils.Error(n.id, "receiveSnapshotMessage", n.color)
	utils.Info(n.id, "receiveSnapshotMessage", "Received snapshot message")
	if (msg.Color == "red" && n.color == "white") {
		utils.Error(n.id, "receiveSnapshotMessage", "SNAPSHOT MODE")

		n.color = "red"
		n.globalStates.PushBack(n.state.ToString())
		msg := Message{
			From:        n.id,
			To:          -1,
			Content:     n.state.ToString(),
			Stamp:       n.clock,
			MessageType: "StateMessage",
			VectClock: 	 n.state.VectClock,
			Color: 		 n.color,
		}
		utils.Info(n.id, "receiveSnapshotMessage", "Send state message")
		go n.writeMessage(msg)
	}
	if (msg.Color == "white" && n.color == "red") {
		utils.Info(n.id, "receiveSnapshotMessage", "Change to prepost message")
		msg.MessageType = "PrepostMessage"
		go n.writeMessage(msg)
	}
  }

  func (n *Net) receiveStateMessage(msg Message) {
	utils.Info(n.id, "receiveStateMessage", "Received state message")
	state := StateFromString(msg.Content)
	if (n.initator) {
		n.globalStates.PushBack(state.ToString())
		n.nbExpectedStates -= 1
		n.nbExpectedMessages += state.Review

		str := fmt.Sprintf("%d states and %d prepost messages to wait to finish the snapshot", n.nbExpectedStates, n.nbExpectedMessages)
		utils.Info(n.id, "receiveStateMessage", str)

		if (n.nbExpectedStates == 0 && n.nbExpectedMessages == 0) {
			utils.Warning(n.id, "receiveStateMessage", "Finish snapshot")

			file, fileErr := os.Create("snapshot.txt")
			if fileErr != nil {
				fmt.Println(fileErr)
				return
			}

			for i := n.globalStates.Front(); i != nil; i = i.Next() {
				fmt.Fprintf(file, "%+v\n", i.Value)
			}

			n.initator = false
		}
	} else {
		utils.Info(n.id, "receiveStateMessage", "not initiator, msg resend")
		go n.writeMessage(msg)
	}
  }

  func (n *Net) receivePrepostMessage(msg Message) {
	utils.Info(n.id, "receivePrepostMessage", "Received prepost message")
	if (n.initator) {
		n.nbExpectedMessages -= 1

		n.globalStates.PushBack(msg.ToString())

		str := fmt.Sprintf("%d states and %d prepost messages to wait to finish the snapshot", n.nbExpectedStates, n.nbExpectedMessages)
		utils.Info(n.id, "receivePrepostMessage", str)

		if (n.nbExpectedStates == 0 && n.nbExpectedMessages == 0) {
			utils.Warning(n.id, "receivePrepostMessage", "Finish snapshot")

			file, fileErr := os.Create("snapshot.txt")
			if fileErr != nil {
				fmt.Println(fileErr)
				return
			}

			for i := n.globalStates.Front(); i != nil; i = i.Next() {
				fmt.Fprintf(file, "%+v\n", i.Value)
			}

			n.initator = false
		}
	} else {
		utils.Info(n.id, "receivePrepostMessage", "not initiator, msg resend")
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
				// process and spread the message on the ring if it doesn't concern the snapshot
				if msg.MessageType != "StateMessage" && msg.MessageType != "PrepostMessage" {
					utils.Info(n.id, "MessageHandler", "Handler processing and sending the message")
					msg.Send()
				}
				n.receiveExternalMessage(msg)
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
	n.globalStates.PushBack(n.state.ToString()) 
	n.nbExpectedStates = NB_SITES - 1
	n.nbExpectedMessages = n.state.Review

	utils.Error(n.id, "InitSnapshot", "Snapshot initialisation")
	str := fmt.Sprintf("%d states and %d prepost messages to wait to finish the snapshot", n.nbExpectedStates, n.nbExpectedMessages)
	utils.Info(n.id, "receivePrepostMessage", str)

	msg := Message{
		From:        n.id,
		To:          -1,
		Content:     "",
		Stamp:       n.clock,
		MessageType: "SnapshotMessage",
		VectClock: 	 n.state.VectClock,
		Color: 		 n.color,
	}

	// data:= fmt.Sprintf("%+v", msg)
	// utils.Info(n.id, "InitSnapshot", data)

	n.writeMessage(msg)
}