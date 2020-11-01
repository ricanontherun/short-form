package logging

import (
	"fmt"
	"os"
	"sync"
)

var allowDebugLogs = false
var once sync.Once

func Debug(msg string) {
	once.Do(func() {
		if _, set := os.LookupEnv("SHORT_FORM_DEBUG"); set {
			allowDebugLogs = true
		}
	})

	if allowDebugLogs {
		fmt.Println("DEBUG: " + msg)
	}
}
