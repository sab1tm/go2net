package server

import "net"

type Server struct {
	Address string
	Listener net.Listener
}

func NewServer(address string) *Server {
	return &Server{
		Address: address,
	}
}

func (s *Server) Start() error {
	var err error
	s.Listener, err = net.Listen("tcp", s.Address)
	if err != nill {
		log.Printf("Error accept: %v", err)
		return err
	}

	log.Printf("Server started")
	go s.acceptLoop()
	select{}
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			log.Printf("failed accept client: %v", err)
		}
		log.Printf("Welcome, %s", conn.RemoteAddr().String())
	}
}
