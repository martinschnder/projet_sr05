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

	// var data = make([]string, 30)
	// data[0] = "Hello World!"
	// var state = NewState(0, data, 3)

	// my_list := list.New()
	// my_list.PushFront(*state)

	// file, fileErr := os.Create("snapshot.txt")
	// if fileErr != nil {
	// 	fmt.Println(fileErr)
	// 	return
	// }

	// for i := my_list.Front(); i != nil; i = i.Next() {
	// 	fmt.Fprintf(file, "%+v\n", i.Value)
	// }
}
