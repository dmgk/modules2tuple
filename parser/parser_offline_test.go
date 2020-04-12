// +build !online

package parser

import "github.com/dmgk/modules2tuple/config"

func init() {
	config.Offline = true
}
