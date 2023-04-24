package main

import (
	"fmt"
	"sync"
	"time"
	"strings"
	"flag"
	"log"
	"os"
)

func ecrivain(timer int) {
	var nom = os.Getpid()
	var h int = 0
	
	for {
		mutex.Lock()
        fmt.Printf(",=sender=%v,=hlg=%d\n", nom, h)
        h = h+1
		mutex.Unlock()
		time.Sleep(time.Duration(timer) * time.Second)
	}
}

func lecteur() {
	var rcvmsg string
	l := log.New(os.Stderr, "", 0)

    for {
        fmt.Scanln(&rcvmsg)

		mutex.Lock()
        tab_allkeyval := strings.Split(rcvmsg[1:], rcvmsg[0:1])

        for _, keyval := range tab_allkeyval {
            tab_keyval := strings.Split(keyval[1:], keyval[0:1])
            l.Println(tab_keyval)
            l.Println("  key : ", tab_keyval[0], " val : ", tab_keyval[1])
        }
		mutex.Unlock()

		rcvmsg = ""
    }
}

var mutex = &sync.Mutex{}

func main() {
	timer := flag.Int("time", 2, "nombre")
	flag.Parse()

	go ecrivain(*timer)
	go lecteur()
	for {
		time.Sleep(time.Duration(60) * time.Second)
	} // Pour attendre la fin des goroutines...
}
