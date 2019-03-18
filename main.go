package main

import (
	lib "./lib"
	"bufio"
	"fmt"
	"net"
	"strings"
)

const port = ":8001"

var memory map[string]string

func main() {
	fmt.Println("Awaiting connections...")

	memory = make(map[string]string)

	listener, err := net.Listen("tcp4", port)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(connection)
	}
}

func handleConnection(connection net.Conn) {
	for {
		netData, err := bufio.NewReader(connection).ReadString('\n')
		if err != nil {
			lib.RespondWithError(connection, "Protocol error: "+err.Error())
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

		commandSplit, err := lib.RequestParser(command)
		if err != nil {
			lib.RespondWithError(connection, err.Error())
			return
		}

		lib.Respond(connection, commandSplit, memory)
	}
	connection.Close()
}
