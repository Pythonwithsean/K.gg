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

type Store interface {
	get(key string) string
	put(key, value string)
	del(key string)
}

type InMemoryStore struct {
	store map[string]string
	mu    *sync.RWMutex
}

type FoundResult struct {
	Found bool
	Index int
}

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

type Server struct {
	addr  string
	port  string
	store *InMemoryStore
}

func NewServer(addr, port string) *Server {
	return &Server{
		addr:  addr,
		port:  port,
		store: NewInMemoryStore(),
	}
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Split(line, " ")
		if len(parts) == 0 {
			continue
		}
		cmd := strings.ToUpper(parts[0])
		switch cmd {
		case "GET":
			log.Printf("[info] handling a connection %s request!\n", cmd)
			found := FoundResult{false, -1}
			for i, part := range parts {
				if strings.Contains(part, "?key=") {
					found.Found = true
					found.Index = i
				}
			}

			if !found.Found || found.Index < 0 {
				continue
			}
			p := strings.Split(parts[found.Index], "=")
			key := p[len(p)-1]
			if value, ok := server.store.get(key); ok {
				conn.Write([]byte(fmt.Sprintf("%s\n", value)))
			} else {
				conn.Write([]byte(fmt.Sprintf("[err] %s not found\n", key)))
			}
		case "POST":
			log.Println("[POST]")
		case "DELETE":
			log.Println("[DEL]")

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
