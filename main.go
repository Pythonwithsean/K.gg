package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const ADDR = "localhost"
const PORT = "8080"

type Store interface {
	Get(key string) string
	Put(key, value string)
	Del(key string)
}

type InMemoryStore struct {
	store map[string]string
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{store: make(map[string]string)}
}

func (s *InMemoryStore) Get(key string) (string, bool) {
	if value, ok := s.store[key]; ok {
		return value, true
	}
	return "", false
}

func (s *InMemoryStore) Put(key, value string) {
	s.store[key] = value
}

func (s *InMemoryStore) Del(key string) {
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
	log.Println("[info] handling a connection")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		parts := strings.Split(line, " ")
		fmt.Println(parts[0])
	}
}

func (server *Server) Start() {
	fullAddr := fmt.Sprintf("%s:%s", server.addr, server.port)
	listener, err := net.Listen("tcp", fullAddr)
	if err != nil {
		log.Fatal("[err]: failed to start server listener")
	}
	log.Printf("Server listening on %s", fullAddr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("[err] accepting connection:", err)
			continue
		}
		server.handleConnection(conn)
	}
	defer listener.Close()
}

func main() {

	server := NewServer(ADDR, PORT)
	server.Start()

}
