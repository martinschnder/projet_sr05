package net

import (
	"container/list"
	"fmt"
	"log"
	"os"
	utils "projet/utils"
  . "projet/message"
)

const NB_SITES = 3

type net struct {
  id int
  clock int
  tab [NB_SITES]Request
  messages list.List
}


func (n net)New(id int) {
  n.id = id
  n.clock = 0
  var tab [NB_SITES]Request
  n.tab = tab
  n.messages = *list.New()
}

func (n net)receiveCSrequest() {
  n.clock += 1
  n.tab[n.id] = Request{
    RequestType: "access",
    Stamp: n.clock,
  }
  msg := Message{
    From: n.id,
    To: -1,
    Content: "",
    Stamp: n.clock,
    MessageType: "LockRequestMessage",
  }
  n.writeMessage(msg)
}

func (n net)receiveCSend() {
  n.clock += 1
  n.tab[n.id] = Request{
    RequestType: "release",
    Stamp: n.clock,
  }
  msg := Message{
    From: n.id,
    To: -1,
    Content: "",
    Stamp: n.clock,
    MessageType: "ReleaseMessage",
  }
  n.writeMessage(msg)
}

func (n net)receiveExternalMessage(msg Message) {
  switch msg.MessageType {
  case "LockRequestMessage":
    n.receiveRequestMessage(msg)
  case "ReleaseMessage":
    n.receiveReleaseMessage(msg)
  case "AckMessage":
    n.receiveAckMessage(msg)
  } 
}

func (n net)receiveRequestMessage(received_msg Message) {
  n.clock = utils.Max(n.clock, received_msg.Stamp) + 1
  n.tab[received_msg.From] = Request{
    RequestType: "access",
    Stamp: received_msg.Stamp,
  }
  msg := Message{
    From: n.id,
    To: received_msg.From,
    Content: "",
    Stamp: n.clock,
    MessageType: "AckMessage",
  }
  n.writeMessage(msg)
  if n.tab[n.id].RequestType == "access" && n.isLastRequest() {
    // TODO send okCS to base
    n.enterCS()
  }
}

func (n net)isLastRequest() bool {
  for i:=0; i<NB_SITES; i++ {
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

func (n net)receiveReleaseMessage(msg Message) {
  n.clock = utils.Max(n.clock, msg.Stamp) + 1
  n.tab[msg.From] = Request{
    RequestType: "release",
    Stamp: msg.Stamp,
  }
  if n.tab[n.id].RequestType == "access" && n.isLastRequest() {
    // TODO send okCS to base
    n.enterCS()
  }
}

func (n net)receiveAckMessage(msg Message) {
  n.clock = utils.Max(n.clock, msg.Stamp) + 1
  if n.tab[msg.From].RequestType == "release" {
    n.tab[msg.From] = Request{
      RequestType: "ack",
      Stamp: msg.Stamp,
    }
  }
  if n.tab[n.id].RequestType == "access" && n.isLastRequest() {
    // TODO send okCS to base
    n.enterCS()
  }
}

func (n net)enterCS() {
  n.logger("Entry in the critical section")
  // TODO start and wait for critical section
  n.basCSRelease()
}

func (n net)basCSRelease() {
  n.clock += 1
  n.logger("BAS CS release received")
  msg := Message{
    From: n.id,
    To: -1,
    Content: "",
    Stamp: n.clock,
    MessageType: "ReleaseMessage",
  }
  n.writeMessage(msg)
}

func (n net)readMessage() {
  var raw string
  fmt.Scanln(&raw)
  var msg = MessageFromString(raw)
  if msg.From != n.id {
    n.messages.PushBack(MessageWrapper{
      Action: "process",
      Message: msg,
    })
  }
}

func (n net)writeMessage(msg Message) {
  n.messages.PushBack(MessageWrapper{
    Action: "send",
    Message: msg,
  })
}

func (n net)logger(content string) {
	stderr := log.New(os.Stderr, "", 0)
  stderr.Printf("From site %d : %s\n", n.id, content)
}

func (n net)messageHandler() {
  for {
    e := n.messages.Front()
    if e != nil { // TODO interface for queue
      wrapperItem := MessageWrapper(e.Value.(MessageWrapper))
      if wrapperItem.Action == "send" {
        n.logger("The handler spreads message on the ring")
        // TODO send message to stdout
      } else if wrapperItem.Action == "process" {
        var msg = wrapperItem.Message
        if msg.To == n.id {
          n.logger("The handler processes message")
          n.receiveExternalMessage(msg)
        } else if msg.To == -1 && msg.From != n.id { // message for all
          // process and spread the message on the ring
          n.logger("The handler processes and spreads message on the ring")
          n.receiveExternalMessage(msg)
          // TODO send message to stdout
        } else { // message not for us 
          // forward the message
          n.logger("The handler forwards the message")
          // TODO send message to stdout
        }

      }
    }
    n.messages.Remove(e)
  }
}
