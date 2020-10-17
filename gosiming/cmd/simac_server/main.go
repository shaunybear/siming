package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	pb "shaunybear/gosiming/internal/simac"

	grpc "google.golang.org/grpc"
)

const (
	port = ":50051"
)

// server is used to implement SiMac server
type server struct {
	pb.UnimplementedSiMacServer
}

// Get Mac Details
func (s *server) GetMacDetails(stream pb.SiMac_GetMacDetailsServer) error {
	fmt.Println("Server: GetMacDetails")
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			fmt.Println("Server: io.EOF")
			return nil
		}
		if err != nil {
			fmt.Printf("Server: err=%v\n", err)
			return err
		}

		fmt.Printf("Server: in=%v\n", in)

		details := &pb.MacDetails{Deveui: &pb.DevEui{Deveui: in.Deveui}, Reply: &pb.MacReply{Status: pb.MacReplyStatus_SUCCESS, ErrorString: "Hi There"}}

		if err := stream.Send(details); err != nil {
			fmt.Printf("Server: Send err=%v\n", err)
			return err
		}
	}
}

func (s *server) Create(context.Context, *pb.Commissioning) (*pb.MacReply, error) {
	return nil, errors.New("Create not implemented")
}

func (s *server) Delete(context.Context, *pb.DevEui) (*pb.MacReply, error) {
	return nil, errors.New("Delete not implemented")
}

func (s *server) Configure(pb.SiMac_ConfigureServer) error {
	return errors.New("Configure")
}

func (s *server) Join(pb.SiMac_JoinServer) error {
	return errors.New("Join not implemented")

}

func (s *server) SendUplink(pb.SiMac_SendUplinkServer) error {
	return errors.New("SendUplink not implemented")
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("gRPC server failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSiMacServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
