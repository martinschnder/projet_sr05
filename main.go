package main

import (
	"flag"
	. "projet/net"
	"time"
)

func main() {
	siteId := flag.Int("id", 0, "Id of the local site")
	port := flag.String("port", "4444", "n° de port")
	addr := flag.String("addr", "localhost", "nom/adresse machine")
	flag.Parse()
	// Initialisation des varibles de chaque site
	var app = NewNet(*siteId, *port, *addr)
	//Lancement de 2 goroutines

	//La goroutine ReadMessage s'occupe de lire les messages qui arrivent sur stdin, et les envoie sur le channel de communication pour MessageHandler
	go app.ReadMessage()

	//La goroutine MessageHandler traite les messages reçus (via l'anneau ou via le client)
	go app.MessageHandler()
	for {
		time.Sleep(time.Duration(60) * time.Second)
	} // Pour attendre la fin des goroutines...
}
