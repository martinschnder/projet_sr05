package projet

import "container/list"

const NB_SITES = 3

type message struct {
  from int
  to int
  content string
  stamp int
}

type request struct {
  type string
  stamp int
}

type net struct {
  id int
  clock int
  tab request[NB_SITES]
  messages list.List
}

type messageWrapper struct {
  action string
  message message
}

func (n net)New(id int) {
  n.id = id
  n.clock = 0
  n.tab = make(request, NB_SITES)
  n.queue = list.New()
}

func (n net)receiveCSrequest() {
  n.clock += 1
  n.tab[n.id] = request("access", n.clock)
  // TODO send request message to all sites
}

func (n net)receiveCSend() {
  n.clock += 1
  n.tab[n.id] = request("freeing", n.clock)
  // TODO send free message to all sites
}


func (n net)receiveRequestMessage(msg message) {
  n.clock = max(n.clock, msg.stamp) + 1
  n.tab[msg.from] = request("access", msg.stamp)
  // TODO send ack message to the sender
  if n.tab[n.id].type == "access" && isLastRequest() {
    // TODO send okCS to base
  }
}

func (n net)isLastRequest() {
  for i := 0; i++; i < NB_SITES {
    if i != n.id {
      if n.tab[i].stamp < n.tab[n.id].stamp {
        return false
      } else if n.tab[i].stamp == n.tab[n.id].stamp {
        if i < n.id {
          return false
        }
      }
    }
  }
  return true
}

func (n net)receiveFreeingMessage(msg message) {
  n.clock = max(n.clock, msg.stamp) + 1
  n.tab[msg.from] = request("freeing", msg.stamp)
  if n.tab[n.id].type == "access" && isLastRequest() {
    // TODO send okCS to base
  }
}

func (n net)receiveAckMessage(msg message) {
  n.clock = max(n.clock, msg.stamp) + 1
  if n.tab[msg.from].type == "freeing" {
    n.tab[msg.from] = request("ack", msg.stamp)
  }
  if n.tab[n.id].type == "access" && isLastRequest() {
    // TODO send okCS to base
  }
}

func (n net)readMessage() {
  fmt.Scanln(&raw)
  msg = Message.fromString(raw) // TODO
  if msg.from != n.id {
    n.messages.InsertAfter(messageWrapper("process", msg))
  }
}
