package main

import (
	"flag"
  "time"
	. "projet/net"
)

func main() {
	siteId := flag.Int("siteId", 0, "Id of the local site")
	flag.Parse()
  var app = *NewNet(*siteId)
  go app.ReadMessage()
  go app.MessageHandler()
	for {
		time.Sleep(time.Duration(60) * time.Second)
	} // Pour attendre la fin des goroutines...
}
