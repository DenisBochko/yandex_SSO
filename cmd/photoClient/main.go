package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	ssov1 "gitlab.crja72.ru/golang/2025/spring/course/projects/go6/contracts/gen/go/sso"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()

	client := ssov1.NewUsersClient(conn)

	data, err := os.ReadFile("cmd/photoClient/saharoza.jpeg")
	if err != nil {
		log.Fatalf("could not read file: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println(data)

	resp, err := client.UploadAvatar(ctx, &ssov1.UploadAvatarRequest{
		UserId:      "4b1a44bb-6f12-4f2a-a7ef-5fdea2d5b22f",
		Photo:       data,
		ContentType: "image/jpeg",
	})
	if err != nil {
		log.Fatalf("upload error: %v", err)
	}

	log.Printf("URL: %s", resp.Url)
}
