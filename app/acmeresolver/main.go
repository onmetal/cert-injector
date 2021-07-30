package main

import (
	"log"
	"os"

	"github.com/onmetal/injector/app/acmeresolver/server"
)

func main() {
	s := server.New()
	if err := s.Run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
