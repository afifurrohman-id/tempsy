package main

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/afifurrohman-id/tempsy/internal/files/models"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"google.golang.org/grpc"
)

type Server struct {
	models.UnimplementedGreeterServer
}

func (server *Server) SayHello(ctx context.Context, req *models.HelloRequest) (*models.HelloReply, error) {
	name := req.GetName()
	log.Println("Received:", name)

	return &models.HelloReply{Message: "Hello " + name}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":"+os.Getenv("PORT"))
	utils.Check(err)

	server := grpc.NewServer()

	models.RegisterGreeterServer(server, new(Server))
	log.Println("server listening at:", lis.Addr())

	utils.Check(server.Serve(lis))

}
