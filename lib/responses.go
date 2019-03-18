package lib

import (
	"fmt"
	"net"
	"strings"
)

func Respond(connection net.Conn, commandSplit []string, memory map[string]string) {
	baseCommand := strings.ToLower(commandSplit[0])

	if baseCommand == "set" {
		HandleSet(connection, commandSplit, memory)
	} else if baseCommand == "get" {
		HandleGet(connection, commandSplit, memory)
	} else if baseCommand == "del" {
		HandleDel(connection, commandSplit, memory)
	} else if baseCommand == "ping" {
		HandlePing(connection, commandSplit)
	} else if len(baseCommand) > 0 {
		RespondWithError(connection, fmt.Sprintf("unknown command '%s'", baseCommand))
	}
}

func HandleSet(connection net.Conn, commandSplit []string, memory map[string]string) {
	if len(commandSplit) == 2 {
		RespondWithError(connection, "wrong number of arguments for 'set' command")
	} else {
		memory[commandSplit[1]] = commandSplit[2]
		RespondPlain(connection, "+OK")
	}
}

func HandleGet(connection net.Conn, commandSplit []string, memory map[string]string) {
	if len(commandSplit) != 2 {
		RespondWithError(connection, "wrong number of arguments for 'get' command")
	} else {
		value, ok := memory[commandSplit[1]]

		if ok {
			RespondWithValue(connection, value)
		} else {
			RespondPlain(connection, "$-1")
		}
	}
}

func HandleDel(connection net.Conn, commandSplit []string, memory map[string]string) {
	if len(commandSplit) == 1 {
		RespondWithError(connection, "wrong number of arguments for 'del' command")
	} else {
		numberOfDeletes := 0

		for i := 1; i < len(commandSplit); i++ {
			_, ok := memory[commandSplit[i]]
			if ok {
				numberOfDeletes += 1
			}
			delete(memory, commandSplit[i])
		}
		RespondPlain(connection, fmt.Sprintf(":%d", numberOfDeletes))
	}
}

func HandlePing(connection net.Conn, commandSplit []string) {
	if len(commandSplit) > 2 {
		RespondWithError(connection, "wrong number of arguments for 'ping' command")
	} else if len(commandSplit) == 2 {
		RespondWithValue(connection, commandSplit[1])
	} else {
		RespondPlain(connection, "+PONG")
	}
}

func RespondWithError(connection net.Conn, response string) {
	message := fmt.Sprintf("-ERR %s\n", response)
	connection.Write([]byte(message))
}

func RespondWithValue(connection net.Conn, response string) {
	message := fmt.Sprintf("$%d\n", len(response))
	connection.Write([]byte(message))
	RespondPlain(connection, response)
}

func RespondPlain(connection net.Conn, response string) {
	connection.Write([]byte(response + "\n"))
}
