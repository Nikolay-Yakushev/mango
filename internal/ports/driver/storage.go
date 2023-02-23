package ports

import (
	"context"

	"github.com/Nikolay-Yakushev/mango/internal/domain/entities/users"
)


type Storage interface{
	GetUser(ctx context.Context, login string)(users.User, error)
	SetUser(ctx context.Context, login, password, email string)(users.User, error)
	BlockUser(ctx context.Context, user users.User)(error)
	GetActive()map[string]users.User
	GetBlocked()map[string]users.User
}