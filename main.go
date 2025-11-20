package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const ADDR = "localhost"

const PORT = "8080"

type Server struct {
	addr  string
	port  string
	store *InMemoryStore
}

type storeValue struct {
	value     string
	expiresAt time.Time
}

type command int

type Store interface {
	get(key string) (string, bool)
	put(key, val string)
	del(key string)
	list() map[string]string
	delExpired()
}

type InMemoryStore struct {
	store map[string]storeValue
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
	return &InMemoryStore{store: make(map[string]storeValue), mu: &sync.RWMutex{}}
}

func (s *InMemoryStore) get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.store[key]
	return val.value, ok
}

func (s *InMemoryStore) put(key, val string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	end := now.Add(5 * time.Minute)
	s.store[key] = storeValue{value: val, expiresAt: end}
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
		result[k] = v.value
	}
	return result
}

func (s *InMemoryStore) delExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	amount := 0
	for key, value := range s.store {
		if now.Sub(value.expiresAt) > 0 {
			amount += 1
			delete(s.store, key)
		}
	}
	log.Printf("deleted %d expired keys\n", amount)
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

func (server *Server) removeExpiredKeys() {
	t := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-t.C:
			log.Println("🗑️ Cleaning up expired keys")
			server.store.delExpired()
		default:
			time.Sleep(time.Second * 30)
		}
	}
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	write := func(msg string) {
		conn.Write([]byte(msg))
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch cmd := parseCommand(line); cmd {
		case GET:
			key, ok := parseGetOrDel(line)
			if !ok {
				write("Invalid get command")
				continue
			}
			if val, ok := server.store.get(key); ok {
				write(fmt.Sprintf("%s\n", val))
			} else {
				write(fmt.Sprintf("key %s not found\n", key))
			}
		case PUT:
			key, val, ok := parsePut(line)

			if !ok {
				write("Invalid put command")
				continue
			}
			server.store.put(key, val)
		case DEL:
			key, ok := parseGetOrDel(line)
			if !ok {
				write("Invalid del command")
				continue
			}
			server.store.del(key)
		case LIST:
			all := server.store.list()
			var response strings.Builder
			for k, v := range all {
				response.WriteString(fmt.Sprintf("%s=%s\n", k, v))
			}
			write(response.String())
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
	go server.removeExpiredKeys()
	for {
		conn, err := listener.Accept()

		log.Println("handling a request")

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
