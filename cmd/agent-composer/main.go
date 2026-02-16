package main

import (
	"log"

	"github.com/jinashrouniyar-268/agent-composer/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
