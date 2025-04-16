# Как использовать AuthInterceptor

### Подключение Interceptor’а к gRPC-серверу

```Go
func main() {
	// Создаем interceptor с секретом и разрешёнными маршрутами
	interceptor, err := authinterceptor.NewAuthInterceptor("our-super-secret", []string{"/auth.Auth/Register", "/auth.Auth/Login"})
	if err != nil {
		log.Fatalf("cannot create interceptor: %v", err)
	}

	// Создаем gRPC-сервер с middleware
	server := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.UnaryAuthMiddleware),
	)

	// Регистрируем свой сервис, как обычно
	// pb.RegisterYourServiceServer(server, &yourService{})
	// ...
}
```

### Как использовать данные из контекста в gRPC-методе

```Go
import (
	"context"
	"<модуль>/authinterceptor"
	"log"
)

// type contextKey string

// const (
// 	ContextUserIDKey   contextKey = "user_id"
// 	ContextEmailKey    contextKey = "email"
// 	ContextNameKey     contextKey = "name"
// 	ContextVerifiedKey contextKey = "verified"
// 	ContextAvatarKey   contextKey = "avatar"
// )

func (s *yourService) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.GetProfileResponse, error) {
	// Получаем user_id из контекста
    // Не забываем, что authinterceptor использует кастомные типы, которые выставляются в контекст (прописаны выше)
	userID, ok := ctx.Value(authinterceptor.ContextUserIDKey).(string)
	if !ok {
		return nil, status.Error(codes.Internal, "user ID not found in context")
	}

	log.Println("Запрос от пользователя:", userID)

	// Дальше логика обработки запроса
	return &pb.GetProfileResponse{...}, nil
}
```