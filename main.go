package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var cyan string = "\033[1;36m"
var raz string = "\033[0;00m"

var h int = 0

func max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func ecrivain(timer int) {
	var nom = os.Getpid()

	for {
		mutex.Lock()
		fmt.Printf(",=sender=%v,=hlg=%d\n", nom, h)
    h++
		mutex.Unlock()
		time.Sleep(time.Duration(timer) * time.Second)
	}
}

func lecteur() {
	var rcvmsg string
	stderr := log.New(os.Stderr, "", 0)

	for {
		fmt.Scanln(&rcvmsg)

		mutex.Lock()
		tab_allkeyval := strings.Split(rcvmsg[1:], rcvmsg[0:1])

    stderr.Printf("%s", cyan)
		for _, keyval := range tab_allkeyval {
			tab_keyval := strings.Split(keyval[1:], keyval[0:1])
      key := tab_keyval[0]
      value := tab_keyval[1]
			// stderr.Println(tab_keyval)
			stderr.Println("  key : ", key, " val : ", value)
      if key == "hlg" {
        h_recu, _ := strconv.ParseInt(value, 10, 16)
        h = max(int(h_recu), h)
      }
		}
    stderr.Printf("%s", raz)
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
