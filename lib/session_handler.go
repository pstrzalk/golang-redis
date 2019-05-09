package lib

import (
	"bufio"
	"fmt"
	"io"
	"net/textproto"
	"strings"

	"github.com/google/shlex"
)

type SessionHandler struct {
	Connection io.ReadWriteCloser
	Memory     Store
	Password   string
	Authorized bool
}

func (s *SessionHandler) HandleConnection() {
	defer s.Connection.Close()
	reader := bufio.NewReader(s.Connection)

	for {
		finish, _ := s.handleSingleConnection(reader)
		if finish {
			break
		}
	}
}

func (s *SessionHandler) handleSingleConnection(reader *bufio.Reader) (bool, error) {
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

func (s *SessionHandler) Respond(commandSplit []string) {
	baseCommand := strings.ToLower(commandSplit[0])

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

func (s *SessionHandler) HandleAuth(commandSplit []string) {
	if len(commandSplit) > 2 {
		s.RespondWithError("wrong number of arguments for 'auth' command")
	} else if commandSplit[1] == s.Password {
		s.Authorized = true
		s.RespondPlain("+OK")
	} else {
		s.RespondWithError("invalid password")
	}
}

func (s *SessionHandler) HandleSet(commandSplit []string) {
	if !s.Authorized {
		s.RespondWithNoAuth()
		return
	}

	if len(commandSplit) == 2 {
		s.RespondWithError("wrong number of arguments for 'set' command")
	} else {
		s.Memory.Set(commandSplit[1], commandSplit[2])
		s.RespondPlain("+OK")
	}
}

func (s *SessionHandler) HandleGet(commandSplit []string) {
	if !s.Authorized {
		s.RespondWithNoAuth()
		return
	}

	if len(commandSplit) != 2 {
		s.RespondWithError("wrong number of arguments for 'get' command")
	} else {
		value, ok, _ := s.Memory.Get(commandSplit[1])

		if ok {
			s.RespondWithValue(value)
		} else {
			s.RespondPlain("$-1")
		}
	}
}

func (s *SessionHandler) HandleDel(commandSplit []string) {
	if !s.Authorized {
		s.RespondWithNoAuth()
		return
	}

	if len(commandSplit) == 1 {
		s.RespondWithError("wrong number of arguments for 'del' command")
	} else {
		numberOfDeletes := 0

		for i := 1; i < len(commandSplit); i++ {
			_, ok, _ := s.Memory.Get(commandSplit[i])
			if ok {
				numberOfDeletes += 1
				s.Memory.Delete(commandSplit[i])
			}
		}
		s.RespondPlain(fmt.Sprintf(":%d", numberOfDeletes))
	}
}

func (s *SessionHandler) HandlePing(commandSplit []string) {
	if !s.Authorized {
		s.RespondWithNoAuth()
		return
	}

	if len(commandSplit) > 2 {
		s.RespondWithError("wrong number of arguments for 'ping' command")
	} else if len(commandSplit) == 2 {
		s.RespondWithValue(commandSplit[1])
	} else {
		s.RespondPlain("+PONG")
	}
}

func (s *SessionHandler) RespondWithNoAuth() {
	message := "-NOAUTH Authentication required.\n"
	s.Connection.Write([]byte(message))
}

func (s *SessionHandler) RespondWithError(response string) {
	message := fmt.Sprintf("-ERR %s\n", response)
	s.Connection.Write([]byte(message))
}

func (s *SessionHandler) RespondWithValue(response string) {
	message := fmt.Sprintf("$%d\n", len(response))
	s.Connection.Write([]byte(message))
	s.RespondPlain(response)
}

func (s *SessionHandler) RespondPlain(response string) {
	s.Connection.Write([]byte(response + "\n"))
}
