package main

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

type CloseableBuffer struct {
	bytes.Buffer
}

func (b *CloseableBuffer) Close() error {
	return nil
}

func TestHandleSingleConnection(t *testing.T) {
	buffer := CloseableBuffer{}
	s := sessionHandler{connection: &buffer, memory: nil, authorized: false}

	reader := bufio.NewReader(s.connection)

	buffer.WriteString("ping")
	s.handleSingleConnection(reader)
	assert.Equal(t, "-NOAUTH Authentication required.\n", buffer.String())
	buffer.Reset()

	s.authorized = true

	buffer.WriteString("\n")
	ret, err := s.handleSingleConnection(reader)
	assert.Equal(t, false, ret)
  assert.Nil(t, err)

	buffer.WriteString("")
	s.handleSingleConnection(reader)
	assert.Equal(t, "-ERR Protocol error: EOF\n", buffer.String())
	buffer.Reset()

	buffer.WriteString("set \"a b")
	s.handleSingleConnection(reader)
	assert.Equal(t, "-ERR EOF found when expecting closing quote\n", buffer.String())
	buffer.Reset()

	buffer.WriteString("quit")
	s.handleSingleConnection(reader)
	assert.Equal(t, "+OK\n", buffer.String())
}

func TestRespondWithNoAuth(t *testing.T) {
	buffer := CloseableBuffer{}
	s := sessionHandler{connection: &buffer, memory: nil}

	s.RespondWithNoAuth()
	assert.Equal(t, "-NOAUTH Authentication required.\n", buffer.String())
}

func TestHandleAuth(t *testing.T) {
	buffer := CloseableBuffer{}
	s := sessionHandler{connection: &buffer, memory: nil, password: "PASS"}

	s.HandleAuth([]string{"auth", "b", "c"})
	assert.Equal(t, "-ERR wrong number of arguments for 'auth' command\n", buffer.String())
	buffer.Reset()

	s.HandleAuth([]string{"auth", "BADPASS"})
	assert.Equal(t, "-ERR invalid password\n", buffer.String())
	buffer.Reset()

	s.HandleAuth([]string{"auth", "PASS"})
	assert.Equal(t, "+OK\n", buffer.String())
	buffer.Reset()
}

func TestHandleSet(t *testing.T) {
	buffer := CloseableBuffer{}
	memory := make(map[string]string)

	s := sessionHandler{connection: &buffer, memory: memory, authorized: true}

	s.HandleSet([]string{"set", "b"})
	assert.Equal(t, "-ERR wrong number of arguments for 'set' command\n", buffer.String())
	buffer.Reset()

	s.HandleSet([]string{"SET", "key", "val"})
	assert.Equal(t, "+OK\n", buffer.String())
	assert.Equal(t, "val", memory["key"])
	buffer.Reset()
}

func TestHandleGet(t *testing.T) {
	buffer := CloseableBuffer{}
	memory := make(map[string]string)
	memory["key"] = "val"

	s := sessionHandler{connection: &buffer, memory: memory, authorized: true}

	s.HandleGet([]string{"get", "a", "b"})
	assert.Equal(t, "-ERR wrong number of arguments for 'get' command\n", buffer.String())
	buffer.Reset()

	s.HandleGet([]string{"get", "key"})
	assert.Equal(t, "$3\nval\n", buffer.String())
	buffer.Reset()

	s.HandleGet([]string{"get", "wrongkey"})
	assert.Equal(t, "$-1\n", buffer.String())
	buffer.Reset()
}

func TestHandleDel(t *testing.T) {
	buffer := CloseableBuffer{}
	memory := make(map[string]string)
	memory["key"] = "val"

	s := sessionHandler{connection: &buffer, memory: memory, authorized: true}

	s.HandleDel([]string{"del"})
	assert.Equal(t, "-ERR wrong number of arguments for 'del' command\n", buffer.String())
	buffer.Reset()

	s.HandleDel([]string{"del", "key"})
	assert.Equal(t, ":1\n", buffer.String())
	value, err := memory["key"]
	assert.Equal(t, "", value)
	assert.NotNil(t, err)
	buffer.Reset()
}

func TestPing(t *testing.T) {
	buffer := CloseableBuffer{}

	s := sessionHandler{connection: &buffer, memory: nil, authorized: true}

	s.HandlePing([]string{"ping", "a", "b"})
	assert.Equal(t, "-ERR wrong number of arguments for 'ping' command\n", buffer.String())
	buffer.Reset()

	s.HandlePing([]string{"ping", "abc"})
	assert.Equal(t, "$3\nabc\n", buffer.String())
	buffer.Reset()

	s.HandlePing([]string{"ping"})
	assert.Equal(t, "+PONG\n", buffer.String())
	buffer.Reset()
}

func TestRespond(t *testing.T) {
	buffer := CloseableBuffer{}
	memory := make(map[string]string)

	s := sessionHandler{connection: &buffer, memory: memory, password: "PASS", authorized: false}
	s.Respond([]string{"get"})
	assert.Equal(t, "-NOAUTH Authentication required.\n", buffer.String())
	buffer.Reset()

	s.Respond([]string{"auth", "PASS"})
	assert.Equal(t, "+OK\n", buffer.String())
	buffer.Reset()

	s.Respond([]string{"set", "key", "val"})
	assert.Equal(t, "+OK\n", buffer.String())
	buffer.Reset()

	s.Respond([]string{"get", "key"})
	assert.Equal(t, "$3\nval\n", buffer.String())
	buffer.Reset()

	s.Respond([]string{"del", "key"})
	assert.Equal(t, ":1\n", buffer.String())
	buffer.Reset()

	s.Respond([]string{"ping"})
	assert.Equal(t, "+PONG\n", buffer.String())
	buffer.Reset()

	s.Respond([]string{"wrongcommand"})
	assert.Equal(t, "-ERR unknown command 'wrongcommand'\n", buffer.String())
	buffer.Reset()
}
