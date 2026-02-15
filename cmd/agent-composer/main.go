package main

import (
	"log"

	"github.com/YOUR_ORG/agent-composer/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
