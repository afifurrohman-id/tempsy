package main

import (
	"context"
	"log"
	"time"

	"github.com/afifurrohman-id/tempsy/internal/files/models"
	"github.com/afifurrohman-id/tempsy/internal/files/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.Dial("grpc.blackisland-5be6586d.southeastasia.azurecontainerapps.io", grpc.WithTransportCredentials(insecure.NewCredentials()))
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
