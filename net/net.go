package net

import (
	"container/list"
	"fmt"
	"os"
	"projet/display"
	. "projet/utils"
)

const NB_SITES = 3

type Net struct {
	id                 int
	clock              int
	requestTab                [NB_SITES]Request
	messages           chan MessageWrapper
	server             *Server
	color              string
	initator           bool
	nbExpectedStates   int
	nbExpectedMessages int
	state              *State
	globalState       *list.List
}

func NewNet(id int, port string, addr string) *Net {
	n := new(Net)

	// Identifiant
	n.id = id

	// Horloge
	n.clock = 0

	var requestTab [NB_SITES]Request
	n.requestTab = requestTab
	for i := 0; i < NB_SITES; i++ {
		n.requestTab[i] = Request{
			RequestType: "release",
			Stamp:       0,
		}
	}

	// Les serveurs communiquent entre eux et servent d'intermédiaire avec les clients
	n.server = NewServer(port, addr, id, n)
	display.Info(n.id, "NewNet", "Successfully created server instance")
	n.messages = make(chan MessageWrapper)

	// Snapshot
	n.color = "white"
	n.initator = false
	n.nbExpectedStates = 0
	n.nbExpectedMessages = 0
	n.state = NewState(id, n.server.text, NB_SITES)
	n.globalState = list.New()

	return n
}

/** ReceiveCSRequest()
Envoi d'une requete d'accès au document partagé
**/
func (n *Net) ReceiveCSRequest() {
	display.Info(n.id, "ReceiveCSRequest", "Received CS request from server")
	n.clock += 1
	n.requestTab[n.id] = Request{
		RequestType: "access",
		Stamp:       n.clock,
	}

	// Création du message
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

/** receiveCSRelease()
Envoi d'une requete de libération du document partagé
**/
func (n *Net) receiveCSRelease() {
	n.clock += 1
	n.requestTab[n.id] = Request{
		RequestType: "release",
		Stamp:       n.clock,
	}
	display.Info(n.id, "receiveCSRelease", "Release CS release from server")
	msg := Message{
		From:        n.id,
		To:          -1,
		Content:     "",
		Stamp:       n.clock,
		MessageType: "ReleaseMessage",
		VectClock:   n.state.VectClock,
		Color:       n.color,
	}
  display.Warning(n.id, "receiveCSRelease", "Quitting critical section")
	go n.writeMessage(msg)
}

/** receiveExternalMessage()
Envoie le message à la fonction appropiée selon le type de message
**/
func (n *Net) receiveExternalMessage(msg Message) {
	// Incrémentation de l'horloge vectorielle
	n.state.VectClockIncr(msg.VectClock, NB_SITES)

	if msg.Color == "red" && n.color == "white" {
		display.Info(n.id, "receiveSnapshotMessage", fmt.Sprintf("Received snapshot message from %d", msg.From))
		display.Warning(n.id, "receiveSnapshotMessage", "Entering in snapshot mode")

		n.color = "red"
		n.globalState.PushBack(n.state.ToString())
		msg := Message{
			From:        n.id,
			To:          -1,
			Content:     n.state.ToString(),
			Stamp:       n.clock,
			MessageType: "StateMessage",
			VectClock:   n.state.VectClock,
			Color:       n.color,
		}
		display.Info(n.id, "receiveSnapshotMessage", "Sending state message")
		go n.writeMessage(msg)
	}
	if msg.Color == "white" && n.color == "red"  && msg.MessageType != "EndSnapshotMessage" {
		newmsg := Message{
			From:        n.id,
			To:          -1,
			Content:     msg.ToStringForContent(),
			Stamp:       n.clock,
			MessageType: "PrepostMessage",
			VectClock:   n.state.VectClock,
			Color:       n.color,
		}
		display.Info(n.id, "receiveSnapshotMessage", "Sending prepost message")
		go n.writeMessage(newmsg)
	}

	switch msg.MessageType {
	case "LockRequestMessage":
		n.receiveRequestMessage(msg)
	case "ReleaseMessage":
		n.receiveReleaseMessage(msg)
	case "AckMessage":
		n.receiveAckMessage(msg)
	case "EditMessage":
		n.server.SendMessage(msg.Content)
	case "StateMessage":
		n.receiveStateMessage(msg)
	case "PrepostMessage":
		n.receivePrepostMessage(msg)
	case "EndSnapshotMessage":
		n.receiveEndSnapshotMessage()
	case "SnapshotMessage":
		return
  default:
    display.Error(n.id, "receiveExternalMessage", "Unknown message type")
	}
}

/** isValidRequest()
vérifie que la reqûete d'accès du site i est valide et que le site i peut modifier la ressource
**/
func (n *Net) isValidRequest() bool {
	for i := 0; i < NB_SITES; i++ {
		//Cette fonction vérifie que la modification peut être faite au niveau temporel
		if i != n.id {
			if n.requestTab[i].Stamp < n.requestTab[n.id].Stamp {
				// On a trouvé une horloge plus faible, on ne peut pas modifier
        		return false
      		} else if n.requestTab[i].Stamp == n.requestTab[n.id].Stamp {
				if i < n.id {
					// On a trouvé une horloge égale, mais l'ordre des sites fait que la modification est impossible
					return false
				}
      		}
		}
	}
	return true
}

/** receiveRequestMessage()
Traitement d'une demande d'accès, envoie un acquittement
**/
func (n *Net) receiveRequestMessage(received_msg Message) {
	display.Info(n.id, "receiveRequestMessage", fmt.Sprintf("Received request message from %d", received_msg.From))
	n.clock = Max(n.clock, received_msg.Stamp) + 1
	n.requestTab[received_msg.From] = Request{
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
	display.Info(n.id, "receiveRequestMessage", fmt.Sprintf("Sending Ack message to %d", received_msg.From))
	go n.writeMessage(msg)
	if n.requestTab[n.id].RequestType == "access" && n.isValidRequest() {
		display.Warning(n.id, "MessageHandler", "Entering in critical section")
		n.server.SendMessage("OkCs")
	}
}

/** receiveReleaseMessage()
Traitement d'un message de libération de ressource
**/
func (n *Net) receiveReleaseMessage(msg Message) {
	display.Info(n.id, "receiveReleaseMessage", fmt.Sprintf("Received release message from %d", msg.From))
	n.clock = Max(n.clock, msg.Stamp) + 1
	n.requestTab[msg.From] = Request{
		RequestType: "release",
		Stamp:       msg.Stamp,
	}
	if n.requestTab[n.id].RequestType == "access" && n.isValidRequest() {
		display.Warning(n.id, "receiveReleaseMessage", "Entering in critical section")
		n.server.SendMessage("OkCs")
	}
}

/** receiveAckMessage()
Traitement de la réception d'un message d'acquittement
**/
func (n *Net) receiveAckMessage(msg Message) {
	display.Info(n.id, "receiveAckMessage", fmt.Sprintf("Received ack message from %d", msg.From))
	n.clock = Max(n.clock, msg.Stamp) + 1
	//Envoi de l'acquittement si pas déjà fait
	if n.requestTab[msg.From].RequestType != "access" {
		n.requestTab[msg.From] = Request{
			RequestType: "ack",
			Stamp:       msg.Stamp,
		}
	}
	if n.requestTab[n.id].RequestType == "access" && n.isValidRequest() {
		display.Warning(n.id, "receiveAckMessage", "Entering in critical section")
		n.server.SendMessage("OkCs")
	}
}

/** receiveStateMessage()
Réception d'un message d'état (ignoré si le site n'est pas l'initiateur)
**/
func (n *Net) receiveStateMessage(msg Message) {
	display.Info(n.id, "receiveStateMessage", fmt.Sprintf("Received state message from %d", msg.From))
	state := StateFromString(msg.Content)
	if n.initator {
		// Sauvegarde de l'état et calcul du nombre d'états à recevoir et de messages attendus.
		n.globalState.PushBack(state.ToString())
		n.nbExpectedStates -= 1
		n.nbExpectedMessages += state.Review

		str := fmt.Sprintf("%d states and %d prepost messages to wait to finish the snapshot", n.nbExpectedStates, n.nbExpectedMessages)
		display.Info(n.id, "receiveStateMessage", str)

		if n.nbExpectedStates == 0 && n.nbExpectedMessages == 0 {
			display.Warning(n.id, "receiveStateMessage", "Finish snapshot")

			file, fileErr := os.Create("snapshot.txt")
			if fileErr != nil {
				fmt.Println(fileErr)
				return
			}

			for i := n.globalState.Front(); i != nil; i = i.Next() {
				fmt.Fprintf(file, "%+v\n", i.Value)
			}

			n.reinitializeAfterSnapshot()
		}
	} else {
		display.Info(n.id, "receiveStateMessage", "not initiator, resending message")
		go n.writeMessage(msg)
	}
}

/** receivePrepostMessage()
réception d'un message étiqueté par un site comme étant un message prépost. Ignoré si le site i n'est pas initiateur
**/
func (n *Net) receivePrepostMessage(msg Message) {
	display.Info(n.id, "receivePrepostMessage", "Received prepost message")
	if n.initator {
		n.nbExpectedMessages -= 1

		//Enregistrement du message prepost dans l'état.
		n.globalState.PushBack(msg.Content)

		str := fmt.Sprintf("%d states and %d prepost messages to wait to finish the snapshot", n.nbExpectedStates, n.nbExpectedMessages)
		display.Info(n.id, "receivePrepostMessage", str)

		if n.nbExpectedStates == 0 && n.nbExpectedMessages == 0 {
			display.Warning(n.id, "receivePrepostMessage", "Finish snapshot")

			file, fileErr := os.Create("snapshot.txt")
			if fileErr != nil {
				fmt.Println(fileErr)
				return
			}

			for i := n.globalState.Front(); i != nil; i = i.Next() {
				fmt.Fprintf(file, "%+v\n", i.Value)
			}

			n.reinitializeAfterSnapshot()
		}
	} else {
		display.Info(n.id, "receivePrepostMessage", "not initiator, msg resend")
		go n.writeMessage(msg)
	}
}

/** reinitializeAfterSnapshot()
Réinitialisation du site initiateur pour la réalisation d'autres snapshots
**/
func (n *Net) reinitializeAfterSnapshot() {
	n.initator = false
	n.color = "white"
	n.globalState = list.New()
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
	display.Info(n.id, "reinitializeAfterSnapshot", "Send EndSnapshotMessage")
	go n.writeMessage(msg)
}

/** receiveEndSnapshotMessage()
Le snapshot initié par un autre site est terminé
**/
func (n *Net) receiveEndSnapshotMessage() {
	n.color = "white"
	n.globalState = list.New()
}

/** ReadMessage()
Lit les message sur l'entrée standard et les envoie ensuite sur le channel de communication du site pour qu'ils soient traités par MessageHandler().
**/
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

/** writeMessage()
Ajout d'un message sur le canal de communication. Le message sera envoyé par MessageHandler()
**/
func (n *Net) writeMessage(msg Message) {
	n.messages <- (MessageWrapper{
		Action:  "send",
		Message: msg,
	})
}

/** MessageHandler()
Lit les messages du site et les traite.
Gère les envois et les réceptions.
**/
func (n *Net) MessageHandler() {
	for {
		//Lecture des messages en attente de traitement
		wrapperItem := <-n.messages
		var msg = wrapperItem.Message
		if wrapperItem.Action == "send" {
			//Le message est propagé sur l'anneau
			if !msg.ConcernSnapshot() {
				n.state.Review += 1
			}
			// Écriture du message sur la sortie standard
			msg.Send()
		} else if wrapperItem.Action == "process" {
			if msg.To == n.id {
				// Message destiné pour le site i
				n.state.Review -= 1
				n.receiveExternalMessage(msg)
			} else if msg.To == -1 && msg.From != n.id {
				// Message pour tous
				// Traitement du message et on continue de le propager
				if !msg.ConcernSnapshot() {
					msg.Send()
				}
				n.receiveExternalMessage(msg)
			} else {
				// Message non destiné au site i,
				// On le transfère (continue la propagation)
				msg.Send()
			}
		}
	}
}

/** SendMessageFromServer()
Fonction utilisée pour envoyer un message depuis le client.
**/
func (n *Net) SendMessageFromServer(msg Message) {
	go n.writeMessage(msg)
}

/** InitSnapshot()
Fonction appellée par le client pour initier le snapshot à partir du site i.
**/
func (n *Net) InitSnapshot() {
	n.color = "red"
	n.initator = true
	n.globalState.PushBack(n.state.ToString())
	n.nbExpectedStates = NB_SITES - 1
	n.nbExpectedMessages = n.state.Review

	display.Warning(n.id, "InitSnapshot", "Snapshot initialisation")
	str := fmt.Sprintf("%d states and %d prepost messages to wait to finish the snapshot", n.nbExpectedStates, n.nbExpectedMessages)
	display.Info(n.id, "receivePrepostMessage", str)

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
