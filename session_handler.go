package main

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
	"strings"

	"github.com/google/shlex"
)

type sessionHandler struct {
	connection io.ReadWriteCloser
	memory     map[string]string
	password   string
	authorized bool
}

func (s *sessionHandler) handleConnection() {
	defer s.connection.Close()
	reader := bufio.NewReader(s.connection)

	for {
		finish, _ := s.handleSingleConnection(reader)
		if finish {
			break
		}
	}
}

func (s *sessionHandler) handleSingleConnection(reader *bufio.Reader) (bool, error) {
	tpReader := textproto.NewReader(reader)

	netData, err := tpReader.ReadLine()
	if err != nil {
		s.RespondWithError("Protocol error: " + err.Error())
		return false, err
	}

	command := strings.TrimSpace(string(netData))
	if command == "" {
		return false, nil
	}

	commandToLower := strings.ToLower(command)
	if commandToLower == "quit" {
		s.RespondPlain("+OK")
		return true, nil
	}

	commandSplit, err := shlex.Split(command)
	if err != nil {
		s.RespondWithError(err.Error())
		return false, err
	}

	s.Respond(commandSplit)
	return false, nil
}

func (s *sessionHandler) Respond(commandSplit []string) {
	baseCommand := strings.ToLower(commandSplit[0])

	if !s.authorized && baseCommand != "auth" {
		s.RespondWithNoAuth()
	} else {
		if baseCommand == "auth" {
			s.HandleAuth(commandSplit)
		} else if baseCommand == "set" {
			s.HandleSet(commandSplit)
		} else if baseCommand == "get" {
			s.HandleGet(commandSplit)
		} else if baseCommand == "del" {
			s.HandleDel(commandSplit)
		} else if baseCommand == "ping" {
			s.HandlePing(commandSplit)
		} else if len(baseCommand) > 0 {
			s.RespondWithError(fmt.Sprintf("unknown command '%s'", baseCommand))
		}
	}
}

func (s *sessionHandler) HandleAuth(commandSplit []string) {
	if len(commandSplit) > 2 {
		s.RespondWithError("wrong number of arguments for 'auth' command")
	} else if commandSplit[1] == s.password {
		s.authorized = true
		s.RespondPlain("+OK")
	} else {
		s.RespondWithError("invalid password")
	}
}

func (s *sessionHandler) HandleSet(commandSplit []string) {
	if len(commandSplit) == 2 {
		s.RespondWithError("wrong number of arguments for 'set' command")
	} else {
		s.memory[commandSplit[1]] = commandSplit[2]
		s.RespondPlain("+OK")
	}
}

func (s *sessionHandler) HandleGet(commandSplit []string) {
	if len(commandSplit) != 2 {
		s.RespondWithError("wrong number of arguments for 'get' command")
	} else {
		value, ok := s.memory[commandSplit[1]]

		if ok {
			s.RespondWithValue(value)
		} else {
			s.RespondPlain("$-1")
		}
	}
}

func (s *sessionHandler) HandleDel(commandSplit []string) {
	if len(commandSplit) == 1 {
		s.RespondWithError("wrong number of arguments for 'del' command")
	} else {
		numberOfDeletes := 0

		for i := 1; i < len(commandSplit); i++ {
			_, ok := s.memory[commandSplit[i]]
			if ok {
				numberOfDeletes += 1
			}
			delete(s.memory, commandSplit[i])
		}
		s.RespondPlain(fmt.Sprintf(":%d", numberOfDeletes))
	}
}

func (s *sessionHandler) HandlePing(commandSplit []string) {
	if len(commandSplit) > 2 {
		s.RespondWithError("wrong number of arguments for 'ping' command")
	} else if len(commandSplit) == 2 {
		s.RespondWithValue(commandSplit[1])
	} else {
		s.RespondPlain("+PONG")
	}
}

func (s *sessionHandler) RespondWithNoAuth() {
	message := "-NOAUTH Authentication required.\n"
	s.connection.Write([]byte(message))
}

func (s *sessionHandler) RespondWithError(response string) {
	message := fmt.Sprintf("-ERR %s\n", response)
	s.connection.Write([]byte(message))
}

func (s *sessionHandler) RespondWithValue(response string) {
	message := fmt.Sprintf("$%d\n", len(response))
	s.connection.Write([]byte(message))
	s.RespondPlain(response)
}

func (s *sessionHandler) RespondPlain(response string) {
	s.connection.Write([]byte(response + "\n"))
}
