package display

import (
    "log"
    "os"
)

// Codes pour le shell
var rouge string = "\033[1;31m"
var orange string = "\033[1;33m"
var raz string = "\033[0;00m"

var Id = "anonyme";
var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)


func Display_d(where string, what string) {
    stderr.Printf(" + [%.6s %d] %-8.8s : %s\n", Id, pid, where, what)
}

func Display_w(where string, what string) {

    stderr.Printf("%s * [%.6s %d] %-8.8s : %s\n%s", orange, Id, pid, where, what, raz)
}

func Display_e(where string, what string) {
    stderr.Printf("%s ! [%.6s %d] %-8.8s : %s\n%s", rouge, Id, pid, where, what, raz)
}