package dbase

import (
	"context"
	"fmt"
	"strings"
	models "github.com/Nikolay-Yakushev/mango/internal/domain"
	users "github.com/Nikolay-Yakushev/mango/internal/domain/entities/users"
	"github.com/google/uuid"
)

func (db *DbStorage)GetBlocked()map[string]users.User{
	u := make(map[string]users.User)
	return u
}

func (db *DbStorage) BlockUser (ctx context.Context, u users.User) (error){
	return nil
}

func (db *DbStorage)GetActive()map[string]users.User{
	u := make(map[string]users.User)
	return u
}

func (db *DbStorage) GetUser(ctx context.Context, l string)(users.User, error) {
	var(
		id       uuid.UUID
		login    string
		password string
		email    string	
	)

	rows, err := db.client.QueryContext(ctx, "SELECT * from users u where u.login = $1", l)

	if err != nil{
		db.log.Sugar().Errorf("errro occured while getting user. Error %s", err.Error())
		return users.User{}, fmt.Errorf("Get user error %w", err)  
	}
	defer rows.Close()

	for rows.Next(){
		err := rows.Scan(&id, &login, &password, &email)
		if err != nil {
			db.log.Sugar().Errorf("failed to fetch results. Error %s", err.Error())
			return users.User{}, fmt.Errorf("failed to fetch results. Error: %w", err)
		
		}
	}
	u := users.User{
		Id: id,
		Login: login,
		Password: password,
		Email: email,
	}
	return u, nil
}


func (db *DbStorage) SetUser(ctx context.Context, login, password, email string) (users.User, error) {
	id := uuid.New()
	_, err := db.client.QueryContext(
			ctx, "INSERT INTO users (id, login, password, email) VALUES ($1, $2, $3, $4);",
			id, 
			login, 
			password,
			email,
		)
	if err != nil{
		db.log.Sugar().Errorf("Error occured while inserting user. Error %s", err.Error())
		// TODO `USER ALREADY EXIST` error how to handle properly?
		if strings.Contains(err.Error(), "duplicate key"){
			err :=  models.ConflictErr
			return users.User{
				Id: id,
				Login: login,
				Password: password,
				Email: email,
			}, err
		}

		return users.User{}, fmt.Errorf("Error occured %w", err)
	}
	u := users.User{
		Id: id,
		Login: login,
		Password: password,
		Email: email,
	}
	return u, nil 
}