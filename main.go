package main

import (
	"fmt"
	"os"

	"flag"

	"github.com/alistanis/cloudconfig/configlib"
)

var (
	setup bool
)

func init() {
	flag.BoolVar(&setup, "setup", false, "Use this flag to perform initial setup")

	flag.Parse()
}

func main() {
	if setup {
		c, err := configlib.GenerateMetaConfig(os.Stdin, os.Stdout)
		if err != nil {
			configlib.ExitError(err, configlib.Unknown)
		}
		fmt.Println(c)
	}
}
