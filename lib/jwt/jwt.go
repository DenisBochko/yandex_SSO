package jwt

import (
	"fmt"
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

// ValidateToken проверяет валидность JWT токена и возвращает claims
func ValidateToken(tokenString string, appSecret string) (jwt.MapClaims, error) {
	// Парсим токен
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(appSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// Проверяем валидность токена
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

/*
// Пример использования
claims, err := ValidateToken(tokenString, appSecret)
if err != nil {
    // Обработка ошибки
    return
}

// Получаем данные из claims
userID := claims["uid"].(string)
email := claims["email"].(string)
exp := claims["exp"].(float64) // JWT числовые значения возвращаются как float64

// Дальнейшая обработка...
*/

//TODO: Покрыть тестами
