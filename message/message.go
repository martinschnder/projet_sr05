package message

import (
	"fmt"
	"strconv"
	"strings"
)

type Message struct {
	From        int
	To          int
	Content     string
	Stamp       int
	MessageType string
}

func MessageFromString(raw string) Message {
  dict := make(map[string]string)
	keyVals := strings.Split(raw[1:], raw[0:1])
	for _, keyVal := range keyVals {
		tuple := strings.Split(keyVal[1:], keyVal[0:1])
		dict[tuple[0]] = tuple[1]
	}
	from, _ := strconv.Atoi(dict["From"])
	to, _ := strconv.Atoi(dict["To"])
	content := dict["Content"]
	messageType := dict["MessageType"]
	stamp, _ := strconv.Atoi(dict["Stamp"])

	msg := Message{
		from,
		to,
		content,
		stamp,
		messageType,
	}
	return msg
}

func (msg Message) ToString() string {
	formatted_str := fmt.Sprintf(",=From=%d,=To=%d,=Content=%s,=Stamp=%d,=MessageType=%s\n", msg.From, msg.To, msg.Content, msg.Stamp, msg.MessageType)
	return formatted_str
}

func (msg Message) Send() {
  fmt.Printf(msg.ToString())
}

type MessageWrapper struct {
	Action  string
	Message Message
}

type Request struct {
	RequestType string
	Stamp       int
}

type Command struct {
  Line int
  Action string
  Content string
}

func ParseCommand(raw string) Command {
  dict := make(map[string]string)
	keyVals := strings.Split(raw[1:], raw[0:1])
	for _, keyVal := range keyVals {
		tuple := strings.Split(keyVal[1:], keyVal[0:1])
		dict[tuple[0]] = tuple[1]
	}
	line, _ := strconv.Atoi(dict["line"])
	action := dict["action"]
	content := dict["message"]

	command := Command {
		line,
		action,
		content,
	}
	return command
}
