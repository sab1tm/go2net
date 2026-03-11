package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type Peer struct {
	Conn        net.Conn
	ConnectedAt time.Time
}

type Message struct {
	Author string
	Text   string
}

type Server struct {
	Address      string
	Listener     net.Listener
	clients      map[net.Conn]*Peer
	deadClients  []net.Conn
	mu           sync.RWMutex
	messagesChan chan Message
}

func NewServer(address string) *Server {
	return &Server{
		Address:      address,
		messagesChan: make(chan Message, 100),
		clients:      make(map[net.Conn]*Peer),
		deadClients:  make([]net.Conn, 0),
	}
}

func (s *Server) Start() error {
	var err error
	s.Listener, err = net.Listen("tcp", s.Address)
	if err != nil {
		log.Printf("Error accept: %v", err)
		return err
	}

	log.Printf("Server started")

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	go s.acceptLoop(ctx)
	go s.Broadcast(ctx)
	defer func() {
		cancel()
		s.Listener.Close()
		close(s.messagesChan)
	}()
	s.Stop(ctx)
	return nil
}

func (s *Server) acceptLoop(ctx context.Context) {
	for {

		conn, err := s.Listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				log.Println("Accept loop stopped")
				return
			default:
				log.Printf("Failed accept client: %v", err)
			}
		}
		log.Printf("Welcome, %s", conn.RemoteAddr().String())
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	s.registerPeer(conn)
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Connection error: %v", err)
			return
		}

		msg := &Message{
			Author: conn.RemoteAddr().String(),
			Text:   string(buf[:n]),
		}
		s.messagesChan <- *msg
	}
}

func (s *Server) Broadcast(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.messagesChan:
			s.mu.RLock()
			message := fmt.Sprintf("%s: %s\n", msg.Author, msg.Text)
			for _, client := range s.clients {
				s.writeInConnection(client.Conn, message)
			}
			s.mu.RUnlock()
			for _, conn := range s.deadClients {
				s.unregisterPeer(conn)
			}
			s.deadClients = s.deadClients[:0]
		}
	}
}

func (s *Server) writeInConnection(conn net.Conn, message string) {
	_, err := conn.Write([]byte(message))
	if err != nil {
		log.Printf("Failed write message: %v", err)
		s.deadClients = append(s.deadClients, conn)
	}
}

func (s *Server) registerPeer(conn net.Conn) {
	peer := &Peer{
		Conn:        conn,
		ConnectedAt: time.Now(),
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[conn] = peer
}

func (s *Server) unregisterPeer(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	conn.Close()
	delete(s.clients, conn)
	log.Printf("Client disconnected: %s", conn.RemoteAddr().String())
}

func (s *Server) Stop(ctx context.Context) {
	<-ctx.Done()
	for _, client := range s.clients {
		s.unregisterPeer(client.Conn)
	}
}
