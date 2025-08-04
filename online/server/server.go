package server

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/reilabs/trusted-setup/online/api"
)

// CeremonyServer represents a server that handles the trusted setup ceremony by coordinating contributors.
type CeremonyServer struct {
	server  *grpc.Server
	service api.CeremonyServiceServer
}

// New creates a new instance of the ceremony.
//
// service implements handlers for the ceremony protocol.
//
// The function returns a handler to the ceremony server. The handler can later be used to start and stop the ceremony.
func New(service api.CeremonyServiceServer) *CeremonyServer {
	s := grpc.NewServer()
	api.RegisterCeremonyServiceServer(s, service)

	return &CeremonyServer{
		server:  s,
		service: service,
	}
}

// Start initializes the ceremony and starts listening for the incoming contributors' connections.
//
// The server will listen on the address specified by host and TCP port specified by port.
//
// The function is not blocking, it spawns a goroutine that listens for connections in the background.
func (s *CeremonyServer) Start(host string, port int) error {
	hostPort := fmt.Sprintf("%s:%d", host, port)
	listener, err := net.Listen("tcp", hostPort)
	if err != nil {
		return err
	}

	go func() {
		err = s.server.Serve(listener)
		if err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	log.Printf("Server started, waiting for Contributors on %s...\n", hostPort)

	return nil
}

// Stop prevents the server from accepting new connections from contributors.
func (s *CeremonyServer) Stop() {
	s.server.GracefulStop()
}
