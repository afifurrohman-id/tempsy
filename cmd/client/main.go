package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"time"

	"github.com/afifurrohman-id/tempsy/internal/files/models"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	secure := flag.Bool("secure", false, "secure using HTTP/2 with TLS, (default is: false)")
	flag.Parse()
	creds := insecure.NewCredentials()
	if *secure {
		creds = credentials.NewTLS(new(tls.Config))
	}

	conn, err := grpc.Dial(flag.Arg(0), grpc.WithTransportCredentials(creds))
	utils.Check(err)
	defer conn.Close()

	client := models.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 18*time.Second)
	defer cancel()

	res, err := client.SayHello(ctx, &models.HelloRequest{
		Name: "Afif",
	})
	utils.Check(err)

	log.Println("Greet:", res.GetMessage())
}
