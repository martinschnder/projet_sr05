package message

import (
    "strings"
    "fmt"
	"appmodule/display"
)

var fieldsep = "/"
var keyvalsep = "="

func Format(key string, val string) string {
    return fieldsep + keyvalsep + key + keyvalsep + val
}

func Send(msg string) {
	display.Info("msg_send", "émission de " + msg)
    fmt.Print(msg + "\n")
}

func Findval(msg string, key string) string {
    if len(msg) < 4 {
		display.Warning("findval", "message trop court : " + msg);
        return ""
    }

    sep := msg[0:1];
    tab_allkeyvals := strings.Split(msg[1:], sep);
   
    for _, keyval := range tab_allkeyvals {
        if len(keyval) >= 4 {
            equ := keyval[0:1]
            tabkeyval := strings.Split(keyval[1:], equ)
            if tabkeyval[0] == key {
                return tabkeyval[1]
            }
        }
    }
        
    return ""
}