package dto

import (
	"chat-server/password"
	"chat-server/repository"
	"fmt"
)

type NewUser struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type User struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Username  string `json:"username"`
}

func ToUserDto(userModel *repository.UserModel) User {
	user := *userModel

	return User{
		ID:        user.ID.Hex(),
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
	}

}

func ToUserModel(user NewUser) (*repository.UserModel, error) {
	hash, err := password.Hash(user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	return &repository.UserModel{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Username:  user.Username,
		Password:  hash,
	}, nil
}
