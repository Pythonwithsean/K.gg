package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

const ADDR = "localhost"

const PORT = "8080"

type Server struct {
	addr  string
	port  string
	store *InMemoryStore
}

type command int

type Store interface {
	get(key string) string
	put(key, value string)
	del(key string)
}

type InMemoryStore struct {
	store map[string]string
	mu    *sync.RWMutex
}

const (
	GET command = iota
	PUT
	DEL
	UNKNOWN
)

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{store: make(map[string]string), mu: &sync.RWMutex{}}
}

func (s *InMemoryStore) get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if value, ok := s.store[key]; ok {
		return value, true
	}
	return "", false
}

func (s *InMemoryStore) put(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = value
}

func (s *InMemoryStore) del(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, key)
}

func NewServer(addr, port string) *Server {
	return &Server{
		addr:  addr,
		port:  port,
		store: NewInMemoryStore(),
	}
}

func parseCommand(line string) command {
	switch cmd := strings.ToUpper(strings.Split(line, " ")[0]); cmd {
	case "GET":
		return GET
	case "PUT":
		return PUT
	case "DEL":
		return DEL
	default:
		return UNKNOWN
	}
}

func parsePut(line string) (string, string, bool) {
	parts := strings.Split(line, " ")
	if len(parts) != 3 {
		return "", "", false
	}
	return parts[1], parts[2], true

}

func parseGetOrDel(line string) (string, bool) {
	parts := strings.Split(line, " ")
	if len(parts) != 2 {
		return "", false
	}
	return parts[1], true
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch cmd := parseCommand(line); cmd {
		case GET:
			key, ok := parseGetOrDel(line)
			if !ok {
				conn.Write([]byte("Invalid get command"))
				continue
			}
			if value, ok := server.store.get(key); ok {
				conn.Write([]byte(fmt.Sprintf("%s\n", value)))
			} else {
				conn.Write([]byte(fmt.Sprintf("key %s not found\n", key)))
			}
		case PUT:
			key, val, ok := parsePut(line)

			if !ok {
				// handle
				conn.Write([]byte("Invalid put command"))
				continue
			}
			server.store.put(key, val)
		case DEL:
			key, ok := parseGetOrDel(line)
			if !ok {
				// handle
				conn.Write([]byte("Invalid del command"))
				continue
			}
			server.store.del(key)
		case UNKNOWN:
			// unknown command
			log.Println("Unknwon command")
		}
	}
}

func (server *Server) Start() {
	fullAddr := fmt.Sprintf("%s:%s", server.addr, server.port)
	listener, err := net.Listen("tcp", fullAddr)
	if err != nil {
		log.Fatal("[err]: failed to start server listener")
	}
	defer listener.Close()
	log.Printf("Server listening on %s", fullAddr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("[err] accepting connection:", err)
			continue
		}
		go server.handleConnection(conn)
	}
}

func main() {

	server := NewServer(ADDR, PORT)
	server.Start()

}
