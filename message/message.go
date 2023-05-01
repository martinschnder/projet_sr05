package message

import (
	"fmt"
	"projet/utils"
	"strconv"
	"strings"
)

type Message struct {
	From        int
	To          int
	Content     string
	Stamp       int
	MessageType string
	VectClock 	[]int
	Color 		string
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
	vectClock := vectClockToArray(dict["VectClock"])
	color := dict["Color"]

	msg := Message{
		from,
		to,
		content,
		stamp,
		messageType,
		vectClock,
		color,
	}
	return msg
}

func vectClockToString(vectClock []int) string {
	formatted_str := "#"
	for i := 0; i < len(vectClock); i++ {
		formatted_str += strconv.Itoa(vectClock[i]) + "#"
	}
	return formatted_str
}

func vectClockToArray(raw string) []int {
	var vectClock = []int{}

	slices := strings.Split(raw[1:], raw[0:1])
	slices = slices[0:len(slices) - 1]
	
    for _, i := range slices {
        j, err := strconv.Atoi(i)
        if err != nil {
            panic(err)
        }
        vectClock = append(vectClock, j)
    }

	return vectClock
}

func (msg Message) ToString() string {
	formatted_str := fmt.Sprintf("|=From=%d|=To=%d|=Content=%s|=Stamp=%d|=MessageType=%s|=VectClock=%s|=Color=%s\n", msg.From, msg.To, msg.Content, msg.Stamp, msg.MessageType, vectClockToString(msg.VectClock), msg.Color)
	return formatted_str
}

func (msg Message) Send() {
  utils.Info(msg.From, "SendMessage", "Sending message")
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
	content := strings.Replace(dict["message"], "$_", " ", -1)

	command := Command {
		line,
		action,
		content,
	}
	return command
}

func (command Command) ToString() string {
  formatted_content := strings.Replace(command.Content, " ", "$_", -1) 
	formatted_str := fmt.Sprintf("~/line/%d~/action/%s~/message/%s", command.Line, command.Action, formatted_content)
	return formatted_str
}


type State struct {
	Id int
	VectClock []int
	Text []string
	Review int
  }

  func NewState(id int, text []string, nbSites int) *State {
	state := new(State)
	state.Id = id
  	state.VectClock = make([]int, nbSites)
	for i := 0; i < nbSites; i++ {
		state.VectClock[i] = 0
	}
	state.Text = text
	state.Review = 0
  	return state
}

func (s *State)VectClockIncr(otherClock []int, nbSites int) {
	s.VectClock[s.Id] += 1 
	for i := 0; i < nbSites; i++ {
		s.VectClock[i] = utils.Max(s.VectClock[i], otherClock[i])
	}
}

func (s *State) ToString() string {
	  formatted_str := fmt.Sprintf("~/Id/%d~/VectClock/%s~/Text/%s~/Review/%d", s.Id, vectClockToString(s.VectClock), textToString(s.Text), s.Review)
	  return formatted_str
  }


func textToString(text []string) string {
	formatted_str := "#"
	for i := 0; i < len(text); i++ {
		text[i] = strings.Replace(text[i], " ", "$_", -1) 
		formatted_str += text[i] + "#"
	}
	return formatted_str
}


func textFromString(raw string) []string {
	text := strings.Split(raw[1:], raw[0:1])
	text = text[0:len(text) - 1]

	for i := 0; i < len(text); i++ {
		text[i] = strings.Replace(text[i], "$_", " ", -1)
	}

	return text
}

func StateFromString(raw string) State {
	dict := make(map[string]string)
	  keyVals := strings.Split(raw[1:], raw[0:1])
	  for _, keyVal := range keyVals {
		  tuple := strings.Split(keyVal[1:], keyVal[0:1])
		  dict[tuple[0]] = tuple[1]
	  }
	  id, _ := strconv.Atoi(dict["Id"])
	  vectClock := vectClockToArray(dict["VectClock"])
	  text := textFromString(dict["Text"])
	  review, _ := strconv.Atoi(dict["Review"])
  
	  state := State{
		  id,
		  vectClock,
		  text,
		  review,
	  }
	  return state
  }