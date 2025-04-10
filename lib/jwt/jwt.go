package jwt

import (
	"time"
	"yandex-sso/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
)

// NewToken creates new JWT token for given user and app.
func NewToken(user models.User, appSecret string, duration time.Duration) (string, error) { 
    token := jwt.New(jwt.SigningMethodHS256)

    // Добавляем в токен всю необходимую информацию 
	// TODO: Добавить ещё и имя пользователя
    claims := token.Claims.(jwt.MapClaims) 
	claims["uid"] = user.ID
    claims["email"] = user.Email
    claims["exp"] = time.Now().Add(duration).Unix() 
	// claims["app_id"] = app.ID

    // Подписываем токен, используя секретный ключ приложения 
    tokenString, err := token.SignedString([]byte(appSecret))
     if err != nil {
         return "", err
    }

    return tokenString, nil
}

//TODO: Покрыть тестами