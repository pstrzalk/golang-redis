package main

import (
	lib "./lib"
	"bufio"
	"fmt"
	"github.com/google/shlex"
	"log"
	"net"
	"net/textproto"
	"os"
	"strings"
)

func main() {
	pN := portNumber()
	fmt.Printf("Awaiting connections on port %s\n", pN)

	memory := make(map[string]string)

	listener, err := net.Listen("tcp", ":"+pN)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(connection, memory)
	}
}

func portNumber() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	} else {
		return "6379"
	}
}

func handleConnection(connection net.Conn, memory map[string]string) {
	for {
		buffer := bufio.NewReader(connection)
		tpReader := textproto.NewReader(buffer)

		netData, err := tpReader.ReadLine()
		if err != nil {
			lib.RespondWithError(connection, "Protocol error: " + err.Error())
			continue
		}

		command := strings.TrimSpace(string(netData))
		if command == "" {
			continue
		}

		commandToLower := strings.ToLower(command)
		if commandToLower == "quit" {
			lib.RespondPlain(connection, "+OK")
			break
		}

		commandSplit, err := shlex.Split(command)
		if err != nil {
			lib.RespondWithError(connection, err.Error())
			continue
		}

		lib.Respond(connection, commandSplit, memory)
	}
	connection.Close()
}
