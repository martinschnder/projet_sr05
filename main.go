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
	var app = NewNet(*siteId, *port, *addr)
	go app.ReadMessage()
	go app.MessageHandler()
	for {
		time.Sleep(time.Duration(60) * time.Second)
	} // Pour attendre la fin des goroutines...
}
