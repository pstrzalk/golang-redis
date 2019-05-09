package main

import (
	"fmt"
	"log"
	"net"

	"github.com/kelseyhightower/envconfig"
	"github.com/pstrzalk/redis-workshop/lib"
)

type config struct {
	Port int    `envconfig:"PORT" default:"6379"`
	Pass string `envconfig:"PASSWORD" default:""`
}

func main() {
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

		go (&lib.SessionHandler{
			Connection: connection,
			Memory:     lib.NewDynamodbStore(),
			Password:   cfg.Pass,
			Authorized: sessionAuthorized,
		}).HandleConnection()
	}
}
