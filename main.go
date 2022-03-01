package main

import (
	"log"

	"github.com/schollz/croc/v9/src/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		log.Fatalln(err)
	}
}
