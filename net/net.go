package net

import (
	// "container/list"
	"fmt"
	"log"
	"os"
  // "time"
	. "projet/message"
	. "projet/utils"
)

const NB_SITES = 3

type Net struct {
	id       int
	clock    int
	tab      [NB_SITES]Request
	// messages list.List
	messages chan MessageWrapper
  // bas Bas
}

func NewNet(id int) *Net {
  n := new(Net)
	n.id = id
	n.clock = 0
	var tab [NB_SITES]Request
	n.tab = tab
	// n.messages = *list.New()
  n.messages = make(chan MessageWrapper)
  return n
}

func (n Net) receiveCSrequest() {
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
	}
	n.writeMessage(msg)
}

func (n Net) receiveCSRelease() {
	n.clock += 1
	n.tab[n.id] = Request{
		RequestType: "release",
		Stamp:       n.clock,
	}
	n.logger("BAS CS release received")
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
	n.writeMessage(msg)
	if n.tab[n.id].RequestType == "access" && n.isLastRequest() {
		// TODO send okCS to base
	}
}

func (n Net) isLastRequest() bool {
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

func (n Net) receiveReleaseMessage(msg Message) {
	n.clock = Max(n.clock, msg.Stamp) + 1
	n.tab[msg.From] = Request{
		RequestType: "release",
		Stamp:       msg.Stamp,
	}
	if n.tab[n.id].RequestType == "access" && n.isLastRequest() {
		// TODO send okCS to base
	}
}

func (n Net) receiveAckMessage(msg Message) {
  n.logger("Received ack message")
	n.clock = Max(n.clock, msg.Stamp) + 1
	if n.tab[msg.From].RequestType == "release" {
		n.tab[msg.From] = Request{
			RequestType: "ack",
			Stamp:       msg.Stamp,
		}
	}
	if n.tab[n.id].RequestType == "access" && n.isLastRequest() {
		// TODO send okCS to base
	}
}

func (n Net) ReadMessage() {
  n.logger("Read message thread initialization")
	var raw string
  for {
    fmt.Scanln(&raw)
    n.logger("Detected new message")
    var msg = MessageFromString(raw)
    if msg.From != n.id {
      // n.messages.PushBack(MessageWrapper{
      // 	Action:  "process",
      // 	Message: msg,
      // })
      n.messages <- MessageWrapper{
        Action:  "process",
        Message: msg,
      }
      n.logger("Pushed new message on the queue")
    }
  }
}

func (n Net) writeMessage(msg Message) {
	// n.messages.PushBack(MessageWrapper{
	// 	Action:  "send",
	// 	Message: msg,
	// })
	n.messages <- (MessageWrapper{
		Action:  "send",
		Message: msg,
	})
}

func (n Net) logger(content string) {
  cyan := "\033[1;36m"
  raz := "\033[0;00m"
	stderr := log.New(os.Stderr, "", 0)
  stderr.Printf("%s", cyan)
	stderr.Printf("From site %d : %s\n", n.id, content)
  stderr.Printf("%s", raz)
}

func (n Net) MessageHandler() {
	for {
		// e := n.messages.Front()
    wrapperItem := <- n.messages
		// if e != nil { // TODO interface for queue
      n.logger("New message on the queue")
			// wrapperItem := MessageWrapper(e.Value.(MessageWrapper))
      var msg = wrapperItem.Message
			if wrapperItem.Action == "send" {
				n.logger("The handler spreads message on the ring")
        msg.Send()
			} else if wrapperItem.Action == "process" {
				if msg.To == n.id {
					n.logger("The handler processes message")
					n.receiveExternalMessage(msg)
				} else if msg.To == -1 && msg.From != n.id { // message for all
					// process and spread the message on the ring
					n.logger("The handler processes and spreads message on the ring")
					n.receiveExternalMessage(msg)
          msg.Send()
				} else { // message not for us
					// forward the message
					n.logger("The handler forwards the message")
          msg.Send()
				}
      // n.messages.Remove(e)
			}
		// } else {
  //     n.logger("No message on the queue")
  //     time.Sleep(time.Duration(2) * time.Second)
  //   }
	}
}

func (n Net)sendToBas(msg Message) {
  n.logger("Net site send message to bas")
  // n.bas.send(Message)
}

func (n Net)sendMessageFromBas(msg Message) {
  n.logger("Bas send message to network")
  n.writeMessage(msg)
}
