package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc/reflection"

	"google.golang.org/grpc"

	"github.com/ryoya-fujimoto/grpc-testing/example/app/pb"
)

const port = ":8080"

type server struct{}

func (s *server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	log.Println("Create user:", req.GetName())

	return &pb.User{Id: 1, Name: req.GetName()}, nil
}

func (s *server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	log.Println("Get user id: ", req.GetId())

	return &pb.User{Id: req.GetId(), Name: "John Smith"}, nil
}

func main() {
	l, err := net.Listen("tcp", port)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &server{})
	reflection.Register(s)
	log.Println("Listening on", port)
	if err := s.Serve(l); err != nil {
		panic(err)
	}
}
