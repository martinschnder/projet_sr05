package utils

import (
	"log"
	"os"
)

var rouge string = "\033[1;31m"
var orange string = "\033[1;33m"
var raz string = "\033[0;00m"
var cyan string = "\033[1;36m"

var pid = os.Getpid()
var stderr = log.New(os.Stderr, "", 0)

func Info(id int, where string, what string) {
	stderr.Printf("%s + [%d %d] %-8.8s : %s\n%s", cyan, id, pid, where, what, raz)
}

func Warning(id int, where string, what string) {

	stderr.Printf("%s * [%d %d] %-8.8s : %s\n%s", orange, id, pid, where, what, raz)
}

func Error(id int, where string, what string) {
	stderr.Printf("%s ! [%d %d] %-8.8s : %s\n%s", rouge, id, pid, where, what, raz)
}

func Max(x int, y int) int {
	if x < y {
		return y
	} else {
		return x
	}
}
