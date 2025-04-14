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

	resp, err := client.UploadAvatar(ctx, &ssov1.UploadAvatarRequest{
		UserId:      "744ca668-8f96-4452-be6d-708cd86c5390",
		Photo:       data,
		ContentType: "image/png",
	})
	if err != nil {
		log.Fatalf("upload error: %v", err)
	}

	log.Printf("URL: %s", resp.Url)
}
