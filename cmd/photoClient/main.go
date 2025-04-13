package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	ssov1 "github.com/DenisBochko/yandex_contracts/gen/go/sso"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()

	client := ssov1.NewUsersClient(conn)

	data, err := os.ReadFile("cmd/photoClient/image.png")
	if err != nil {
		log.Fatalf("could not read file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println(data)

	resp, err := client.UploadPhoto(ctx, &ssov1.UploadPhotoRequest{
		UserId:      "user456",
		FileName:    "avatar.jpg",
		ContentType: "image/jpeg",
		Photo:       data,
	})
	if err != nil {
		log.Fatalf("upload error: %v", err)
	}

	log.Printf("Status: %s\nMessage: %s\nURL: %s", resp.Status, resp.Message, resp.Url)
}
