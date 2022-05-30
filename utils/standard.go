package DCDUtils

import (
	"fmt"
	"runtime"
	"time"
)

// Add extra line of space when the program exits
func ExitNewLine() string {
	var endOfProgram string
	if runtime.GOOS == "windows" {
		endOfProgram = "\n"
	} else {
		endOfProgram = "\n\n"
	}
	return endOfProgram
}

// Check for root, as running the program with it can change the path
func RootCheck(tuid int) {
	if tuid == 0 {
		fmt.Print("[NOTICE] This program is running as root\n")
		fmt.Print("[...] The logged in user will be used.\n\n")
	}
}

// Set a date and time
func TimeDate() string {
	timeDat := time.Now().Format("2006-01-02--15-04-05")
	return timeDat
}