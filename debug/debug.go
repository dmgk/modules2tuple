package debug

import (
	"fmt"
	"os"

	"github.com/dmgk/modules2tuple/config"
)

func Print(v ...interface{}) {
	if config.Debug {
		fmt.Fprint(os.Stderr, v...)
	}
}

func Printf(format string, v ...interface{}) {
	if config.Debug {
		fmt.Fprintf(os.Stderr, format, v...)
	}
}
