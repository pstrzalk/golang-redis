package main

import (
	"fmt"
	"log"
	"net"

	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Port int    `envconfig:"PORT" default:"6379"`
	Pass string `envconfig:"PASSWORD" default:""`
}

func main() {
	memory := make(map[string]string)

	var cfg config
	envconfig.MustProcess("", &cfg)

	port := fmt.Sprintf(":%d", cfg.Port)
	fmt.Printf("Awaiting connections on port %s\n", port)

	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	sessionAuthorized := cfg.Pass == ""

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go (&sessionHandler{
			connection: connection,
			memory:     memory,
			password:   cfg.Pass,
			authorized: sessionAuthorized,
		}).handleConnection()
	}
}
