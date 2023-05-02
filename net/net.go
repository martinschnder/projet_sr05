package net

import (
	"container/list"
	"fmt"
	"os"
	. "projet/message"
	"projet/utils"
	. "projet/utils"
)

const NB_SITES = 3

type Net struct {
	id                 int
	clock              int
	tab                [NB_SITES]Request
	messages           chan MessageWrapper
	server             *Server
	color              string
	initator           bool
	nbExpectedStates   int
	nbExpectedMessages int
	state              *State
	globalStates       *list.List
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
	utils.Info(n.id, "ReceiveCSRequest", "received CS request from server")
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
		VectClock:   n.state.VectClock,
		Color:       n.color,
	}
	n.writeMessage(msg)
}

func (n *Net) receiveCSRelease() {
	n.clock += 1
	n.tab[n.id] = Request{
		RequestType: "release",
		Stamp:       n.clock,
	}
	utils.Info(n.id, "receiveCSRelease", "Server CS release received")
	msg := Message{
		From:        n.id,
		To:          -1,
		Content:     "",
		Stamp:       n.clock,
		MessageType: "ReleaseMessage",
		VectClock:   n.state.VectClock,
		Color:       n.color,
	}
  utils.Warning(n.id, "receiveCSRelease", "Quitting critical section")
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
	case "EndSnapshotMessage":
		n.receiveEndSnapshotMessage()
  default:
    utils.Error(n.id, "receiveExternalMessage", "Unknown message type")
	}
}

func (n *Net) receiveRequestMessage(received_msg Message) {
	utils.Info(n.id, "receiveRequestMessage", fmt.Sprintf("Received request message from %d", received_msg.From))
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
		VectClock:   n.state.VectClock,
		Color:       n.color,
	}
	utils.Info(n.id, "receiveRequestMessage", fmt.Sprintf("Sending Ack message to %d", received_msg.From))
	go n.writeMessage(msg)
	if n.tab[n.id].RequestType == "access" && n.isValidRequest() {
		utils.Warning(n.id, "MessageHandler", "Entering in critical section")
		n.server.SendMessage("OkCs")
	}
}

func (n *Net) isValidRequest() bool {
	for i := 0; i < NB_SITES; i++ {
		if i != n.id {
      if n.tab[i].Stamp < n.tab[n.id].Stamp {
        return false
      } else if n.tab[i].Stamp == n.tab[n.id].Stamp {
        if i < n.id {
          return false
        }
      }
		}
	}
	return true
}

func (n *Net) receiveReleaseMessage(msg Message) {
	utils.Info(n.id, "receiveReleaseMessage", fmt.Sprintf("Received release message from %d", msg.From))
	n.clock = Max(n.clock, msg.Stamp) + 1
	n.tab[msg.From] = Request{
		RequestType: "release",
		Stamp:       msg.Stamp,
	}
	if n.tab[n.id].RequestType == "access" && n.isValidRequest() {
		utils.Warning(n.id, "receiveReleaseMessage", "Entering in critical section")
		n.server.SendMessage("OkCs")
	}
}

func (n *Net) receiveAckMessage(msg Message) {
	utils.Info(n.id, "receiveAckMessage", fmt.Sprintf("Received ack message from %d", msg.From))
	n.clock = Max(n.clock, msg.Stamp) + 1
	if n.tab[msg.From].RequestType != "access" {
		n.tab[msg.From] = Request{
			RequestType: "ack",
			Stamp:       msg.Stamp,
		}
	}
	if n.tab[n.id].RequestType == "access" && n.isValidRequest() {
		utils.Warning(n.id, "receiveAckMessage", "Entering in critical section")
		n.server.SendMessage("OkCs")
	}
}

func (n *Net) receiveSnapshotMessage(msg Message) {
	utils.Info(n.id, "receiveSnapshotMessage", fmt.Sprintf("Received snapshot message from %d", msg.From))
	if msg.Color == "red" && n.color == "white" {
		utils.Warning(n.id, "receiveSnapshotMessage", "Entering in snapshot mode")

		n.color = "red"
		n.globalStates.PushBack(n.state.ToString())
		msg := Message{
			From:        n.id,
			To:          -1,
			Content:     n.state.ToString(),
			Stamp:       n.clock,
			MessageType: "StateMessage",
			VectClock:   n.state.VectClock,
			Color:       n.color,
		}
		utils.Info(n.id, "receiveSnapshotMessage", "Sending state message")
		go n.writeMessage(msg)
	}
	if msg.Color == "white" && n.color == "red" {
		utils.Info(n.id, "receiveSnapshotMessage", "Change to prepost message")
		msg.MessageType = "PrepostMessage"
		go n.writeMessage(msg)
	}
}

func (n *Net) receiveStateMessage(msg Message) {
	utils.Info(n.id, "receiveStateMessage", fmt.Sprintf("Received state message from %d", msg.From))
	state := StateFromString(msg.Content)
	if n.initator {
		n.globalStates.PushBack(state.ToString())
		n.nbExpectedStates -= 1
		n.nbExpectedMessages += state.Review

		str := fmt.Sprintf("%d states and %d prepost messages to wait to finish the snapshot", n.nbExpectedStates, n.nbExpectedMessages)
		utils.Info(n.id, "receiveStateMessage", str)

		if n.nbExpectedStates == 0 && n.nbExpectedMessages == 0 {
			utils.Warning(n.id, "receiveStateMessage", "Finish snapshot")

			file, fileErr := os.Create("snapshot.txt")
			if fileErr != nil {
				fmt.Println(fileErr)
				return
			}

			for i := n.globalStates.Front(); i != nil; i = i.Next() {
				fmt.Fprintf(file, "%+v\n", i.Value)
			}

			n.reinitializeAfterSnapshot()
		}
	} else {
		utils.Info(n.id, "receiveStateMessage", "not initiator, resending message")
		go n.writeMessage(msg)
	}
}

func (n *Net) receivePrepostMessage(msg Message) {
	utils.Info(n.id, "receivePrepostMessage", "Received prepost message")
	if n.initator {
		n.nbExpectedMessages -= 1

		n.globalStates.PushBack(msg.ToString())

		str := fmt.Sprintf("%d states and %d prepost messages to wait to finish the snapshot", n.nbExpectedStates, n.nbExpectedMessages)
		utils.Info(n.id, "receivePrepostMessage", str)

		if n.nbExpectedStates == 0 && n.nbExpectedMessages == 0 {
			utils.Warning(n.id, "receivePrepostMessage", "Finish snapshot")

			file, fileErr := os.Create("snapshot.txt")
			if fileErr != nil {
				fmt.Println(fileErr)
				return
			}

			for i := n.globalStates.Front(); i != nil; i = i.Next() {
				fmt.Fprintf(file, "%+v\n", i.Value)
			}

			n.reinitializeAfterSnapshot()
		}
	} else {
		utils.Info(n.id, "receivePrepostMessage", "not initiator, msg resend")
		go n.writeMessage(msg)
	}
}

func (n *Net) reinitializeAfterSnapshot() {
	n.initator = false
	n.color = "white"
	n.globalStates = list.New()
	n.nbExpectedStates = 0
	n.nbExpectedMessages = 0
	msg := Message{
		From:        n.id,
		To:          -1,
		Content:     "",
		Stamp:       n.clock,
		MessageType: "EndSnapshotMessage",
		VectClock:   n.state.VectClock,
		Color:       n.color,
	}
	utils.Info(n.id, "reinitializeAfterSnapshot", "Send EndSnapshotMessage")
	go n.writeMessage(msg)
}

func (n *Net) receiveEndSnapshotMessage() {
	n.color = "white"
	n.globalStates = list.New()
}

func (n *Net) ReadMessage() {
	var raw string
	for {
		fmt.Scanln(&raw)
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
	n.messages <- (MessageWrapper{
		Action:  "send",
		Message: msg,
	})
}

func (n *Net) MessageHandler() {
	for {
		wrapperItem := <-n.messages
		var msg = wrapperItem.Message
		if wrapperItem.Action == "send" {
			if !msg.ConcernSnapshot() {
				n.state.Review += 1
			}
			msg.Send()
		} else if wrapperItem.Action == "process" {
			if msg.To == n.id {
				n.state.Review -= 1
				n.receiveExternalMessage(msg)
			} else if msg.To == -1 && msg.From != n.id { // message for all
				// process and spread the message on the ring if it doesn't concern the snapshot
				if !msg.ConcernSnapshot() {
					msg.Send()
				}
				n.receiveExternalMessage(msg)
			} else { // message not for us
				// forward the message
				msg.Send()
			}
		}
		// time.Sleep(time.Duration(1) * time.Second)
	}
}

func (n *Net) SendMessageFromServer(msg Message) {
	go n.writeMessage(msg)
}

func (n *Net) InitSnapshot() {
	n.color = "red"
	n.initator = true
	n.globalStates.PushBack(n.state.ToString())
	n.nbExpectedStates = NB_SITES - 1
	n.nbExpectedMessages = n.state.Review

	utils.Warning(n.id, "InitSnapshot", "Snapshot initialisation")
	str := fmt.Sprintf("%d states and %d prepost messages to wait to finish the snapshot", n.nbExpectedStates, n.nbExpectedMessages)
	utils.Info(n.id, "receivePrepostMessage", str)

	msg := Message{
		From:        n.id,
		To:          -1,
		Content:     "",
		Stamp:       n.clock,
		MessageType: "SnapshotMessage",
		VectClock:   n.state.VectClock,
		Color:       n.color,
	}

	n.writeMessage(msg)
}
