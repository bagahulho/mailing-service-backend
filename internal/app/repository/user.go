package repository

import (
	"RIP/internal/app/ds"
	"errors"
	"fmt"
)

func (r *Repository) CreateUser(user *ds.User) error {
	// Проверяем, существует ли уже пользователь с таким логином
	var existingUser ds.User
	if err := r.db.Where("login = ?", user.Login).First(&existingUser).Error; err == nil {
		return errors.New("пользователь с таким логином уже существует")
	}

	return r.db.Create(user).Error
}

func (r *Repository) UpdateUser(newUser ds.User, id uint) (ds.User, error) {
	var user ds.User
	if err := r.db.First(&user, id).Error; err != nil {
		return user, fmt.Errorf("пользователь с id %d не найден", id)
	}

	if newUser.Login != "" {
		user.Login = newUser.Login
	}
	if newUser.Password != "" {
		user.Password = newUser.Password
	}

	if err := r.db.Save(user).Error; err != nil {
		return user, err
	}
	return user, nil
}

func (r *Repository) AuthUser() {

}

func (r *Repository) DeAuthUser() {

}
