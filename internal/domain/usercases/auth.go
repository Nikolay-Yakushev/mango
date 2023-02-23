package usercases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Nikolay-Yakushev/mango/internal/adapters/memory"
	"github.com/Nikolay-Yakushev/mango/internal/adapters/dbase"
	models "github.com/Nikolay-Yakushev/mango/internal/domain"
	"github.com/Nikolay-Yakushev/mango/internal/domain/entities/users"
	ports "github.com/Nikolay-Yakushev/mango/internal/ports/driver"
	cfg "github.com/Nikolay-Yakushev/mango/pkg/config"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	storage   ports.Storage
	log       *zap.Logger
	cfg 	  *cfg.Config

}
func (a *Auth) hashPassword(password string)(string, error){
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
		return "", err
    }
    return string(hash), nil
}

func (a *Auth) encodePassword (passwrod string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(passwrod), bcrypt.DefaultCost)
    if err != nil {
		return "", err
    }
    return string(hash), nil
}

func(a *Auth) generateToken(login string, ttl time.Duration) (string, error) {
	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		Issuer: login,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedjwt, err := token.SignedString([]byte(a.cfg.Srv.SecretSignature))
	if err != nil{
		return "", err
	}
	return signedjwt, nil
}


func (a *Auth) verifyToken(tokenString string) (string, error){
	var claims jwt.RegisteredClaims

	validateSign := func(token *jwt.Token) (interface{}, error) {
		return []byte(a.cfg.Srv.SecretSignature), nil
	}

	token , err := jwt.ParseWithClaims(tokenString, &claims, validateSign)
	if !token.Valid {
		return "", err
	} else if errors.Is(err, jwt.ErrTokenMalformed) {
		return "", models.TokenInvalidErr
	} else if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet) {
		return "", models.TokenExpiredErr
	}
	return claims.Issuer, nil

}


func New(ctx context.Context, logger *zap.Logger, cfg *cfg.Config) (*Auth, error) {
	// TODO some config to parse. e.g: inmemory | redis | mongo | postgres
	namedLogger := logger.Named("auth")
	var (
		storage ports.Storage
		err error
	)
	if cfg.Dbase != "postgres"{
		storage, err = memory.New(logger)
		
	}else{
		storage, err = dbase.New(ctx, logger)
	}
	
	if err != nil {
		logger.Sugar().Errorw("Memory start failed", "error", err)
		err := fmt.Errorf("Memory start failed. Reason: %w", err)
		return nil, err
	}
	return &Auth{
		storage: storage,
		log: namedLogger,
		cfg: cfg,
	 }, nil
}

func (a *Auth) Login(ctx context.Context, login, password string) (string, string, error)  {
	u, err := a.storage.GetUser(ctx, login)
	if err != nil {
		a.log.Sugar().Errorw("Storage error", "reason", err)
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		a.log.Sugar().Errorw("Password mismatch of user=%s", login)
		return "", "", err
	}
	
	access, err := a.generateToken(login, a.cfg.Srv.AccessTokenExpired)
	if err != nil {
		a.log.Sugar().Errorw("AccessToken generation failed of user=%s", "login", login)
		return "", "", err
	}

	refresh, err := a.generateToken(login, a.cfg.Srv.RefreshTokenExpired)
	if err != nil {
		a.log.Sugar().Errorw("RefreshToken generation failed of user=%s", "login", login)
		return "", "", err
	}

	return access, refresh, nil

}

func (a *Auth) Logout(ctx context.Context, login, password string) (bool, error) { 
	u, err := a.storage.GetUser(ctx, login)
	if err != nil {
		a.log.Sugar().Errorw("Storage error", "reason", err)
		return false, err
	}
	// make sure we block a verified user
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		a.log.Sugar().Errorw("Password mismatch of user=%s", login)
		return false, err
	}

	if err := a.storage.BlockUser(ctx, u); err != nil {
		a.log.Sugar().Errorw("User block failed", "reason", err)
		return false, err
	}
	return true, nil
}

func (a *Auth) Verify(
	ctx context.Context, accessToken, refreshToken string) (users.VerifyResponse, error) {
	var (
		login string
		r users.VerifyResponse
		err error
	)

	login, err = a.verifyToken(accessToken)
	if err != nil && errors.Is(err, models.TokenExpiredErr){
		return r, err
	}

	if err != nil {
		login, err = a.verifyToken(refreshToken)
		if err != nil {
			return r, err
		}
	}
	// check user is blocked
	_, err = a.storage.GetUser(ctx, login)
	if err !=nil {
		return r, err
	}

	r.User, err = a.storage.GetUser(ctx, login)
	if err != nil {
		return r, err
	}
	
	r.AccessToken, err = a.generateToken(login, a.cfg.Srv.AccessTokenExpired)
	if err != nil {
		return r, err
	}

	r.RefreshToken, err = a.generateToken(login, a.cfg.Srv.RefreshTokenExpired)
	if err != nil {
		return r, err
	}
	
	return r, nil
}

func (a *Auth) Singup(ctx context.Context, login, password, email string) (users.User, error)  {
	pswd, err := a.hashPassword(password)
	if err != nil{
		return users.User{}, err
	}
	user, err := a.storage.SetUser(ctx, login, pswd, email)
	if err !=nil {
		a.log.Sugar().Errorw("Storage error", "reason", err)
		return users.User{}, err
	}
	return user, nil
}