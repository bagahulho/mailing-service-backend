package repository

import (
	"RIP/internal/app/ds"
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

func (r *Repository) SaveSession(ctx context.Context, userID uint, token string, expiration time.Duration) error {
	err := r.redisClient.Set(ctx, fmt.Sprintf("session:%d", userID), token, expiration).Err()
	return err
}

func (r *Repository) GetSession(ctx context.Context, userID uint) (string, error) {
	token, err := r.redisClient.Get(ctx, fmt.Sprintf("session:%d", userID)).Result()
	if errors.Is(err, redis.Nil) {
		return "", errors.New("сессия не найдена")
	}
	return token, err
}

func (r *Repository) DeleteSession(ctx context.Context, userID uint) error {
	return r.redisClient.Del(ctx, fmt.Sprintf("session:%d", userID)).Err()
}

func (r *Repository) RegisterUser(login, password string) error {
	var exists ds.User
	if err := r.db.Where("login = ?", login).First(&exists).Error; err == nil {
		return errors.New("пользователь с таким логином уже существует")
	}

	user := ds.User{Login: login, Password: password}
	return r.db.Create(&user).Error
}

func (r *Repository) AuthenticateUser(login, password string) (*ds.User, error) {
	var user ds.User
	// Поиск пользователя по логину
	if err := r.db.Where("login = ?", login).First(&user).Error; err != nil {
		return nil, errors.New("пользователь не найден")
	}

	// Проверка пароля
	if user.Password != password {
		return nil, errors.New("неверный пароль")
	}

	return &user, nil
}

//func (r *repository) UpdateUser(newUser ds.User, id uint) (ds.User, error) {
//	var user ds.User
//	if err := r.db.First(&user, id).Error; err != nil {
//		return user, fmt.Errorf("пользователь с id %d не найден", id)
//	}
//
//	if newUser.Login != "" {
//		user.Login = newUser.Login
//	}
//	if newUser.Password != "" {
//		user.Password = newUser.Password
//	}
//
//	if err := r.db.Save(user).Error; err != nil {
//		return user, err
//	}
//	return user, nil
//}
//
//func (r *repository) AuthUser() {
//
//}
//
//func (r *repository) DeAuthUser() {
//
//}
