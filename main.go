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
	get(key string) (string, bool)
	put(key, val string)
	del(key string)
	list() map[string]string
}

type InMemoryStore struct {
	store map[string]string
	mu    *sync.RWMutex
}

const (
	UNKNOWN command = iota
	GET
	PUT
	DEL
	LIST
)

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{store: make(map[string]string), mu: &sync.RWMutex{}}
}

func (s *InMemoryStore) get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.store[key]
	return val, ok
}

func (s *InMemoryStore) put(key, val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = val
}

func (s *InMemoryStore) del(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, key)
}

func (s *InMemoryStore) list() map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]string)
	for k, v := range s.store {
		result[k] = v
	}
	return result
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
	case "LIST":
		return LIST
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
			if val, ok := server.store.get(key); ok {
				conn.Write([]byte(fmt.Sprintf("%s\n", val)))
			} else {
				conn.Write([]byte(fmt.Sprintf("key %s not found\n", key)))
			}
		case PUT:
			key, val, ok := parsePut(line)

			if !ok {
				conn.Write([]byte("Invalid put command"))
				continue
			}
			server.store.put(key, val)
		case DEL:
			key, ok := parseGetOrDel(line)
			if !ok {
				conn.Write([]byte("Invalid del command"))
				continue
			}
			server.store.del(key)
		case LIST:
			all := server.store.list()
			var response strings.Builder
			for k, v := range all {
				response.WriteString(fmt.Sprintf("%s=%s\n", k, v))
			}
			conn.Write([]byte(response.String()))
		case UNKNOWN:
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
