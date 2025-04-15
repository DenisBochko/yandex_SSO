package storage

import "errors"

var (
    ErrUserExists = errors.New("user already exists") 
	ErrUserNotFound = errors.New("user not found")
    ErrAppNotFound = errors.New("app not found")
    ErrInternalStorage = errors.New("internal storage error")
    ErrTokenExists = errors.New("token already exists")
    ErrTokenNotFound = errors.New("token not found")
    ErrTokenExpired = errors.New("token expired")
)

