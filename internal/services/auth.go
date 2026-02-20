package services

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
	"github.com/spf13/viper"

	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/gateways"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
)

type Tokens struct {
	Access  string
	Refresh string
}

type UserClaims struct {
	jwt.StandardClaims
	Id   uint
	Role models.Role
}

type authService struct {
	userGateway       gateways.UserGateway
	mosquittoGateway  gateways.MosquittoGateway
	accessSigningKey  []byte
	accessTokenTTL    time.Duration
	refreshSigningKey []byte
	refreshTokenTTL   time.Duration
}

func NewAuthService(
	userGateway gateways.UserGateway,
	mosquittoGateway gateways.MosquittoGateway,
) *authService {
	return &authService{
		userGateway:       userGateway,
		mosquittoGateway:  mosquittoGateway,
		accessSigningKey:  []byte(viper.GetString("auth_access_signing_key")),
		accessTokenTTL:    viper.GetDuration("auth_access_token_ttl"),
		refreshSigningKey: []byte(viper.GetString("auth_refresh_signing_key")),
		refreshTokenTTL:   viper.GetDuration("auth_refresh_token_ttl"),
	}
}

func (a *authService) SignUp(newUser models.UserCore) error {
	if !utils.IsValidEmail(newUser.Email) {
		return utils.ResponseError{
			Code:    http.StatusBadRequest,
			Message: consts.ErrIncorrectPasswordOrEmail,
		}
	}

	exist, err := a.userGateway.DoesExistEmail(0, newUser.Email)
	if err != nil {
		return err
	}
	if exist {
		return utils.ResponseError{
			Code:    http.StatusBadRequest,
			Message: consts.ErrEmailAlreadyInUse,
		}
	}

	if len(newUser.Password) < 8 {
		return utils.ResponseError{
			Code:    http.StatusBadRequest,
			Message: consts.ErrShortPassword,
		}
	}

	password := newUser.Password
	passwordHash := utils.HashPassword(password)
	newUser.Password = passwordHash

	if err = a.userGateway.CreateUser(newUser); err != nil {
		return err
	}

	a.mosquittoGateway.WriteMosquittoPasswd(newUser.Email, password)
	a.mosquittoGateway.WriteNewUserToAcl(newUser.Email)
	return nil
}

func (a *authService) SignIn(email, password string) (Tokens, error) {
	user, err := a.userGateway.GetUserByEmail(email)
	if err != nil {
		return Tokens{}, err
	}

	if err = utils.ComparePassword(user.Password, password); err != nil {
		return Tokens{}, utils.ResponseError{
			Code:    http.StatusBadRequest,
			Message: consts.ErrIncorrectPasswordOrEmail,
		}
	}

	access, err := generateToken(user, a.accessTokenTTL, a.accessSigningKey)
	if err != nil {
		return Tokens{}, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	refresh, err := generateToken(user, a.refreshTokenTTL, a.refreshSigningKey)
	if err != nil {
		return Tokens{}, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	return Tokens{Access: access, Refresh: refresh}, nil
}

func (a *authService) Refresh(token string) (string, error) {
	claims, err := parseToken(token, a.refreshSigningKey)
	if err != nil {
		return "", err
	}

	user := models.UserCore{
		ID:   claims.Id,
		Role: claims.Role,
	}

	newAccessToken, err := generateToken(user, a.accessTokenTTL, a.accessSigningKey)
	if err != nil {
		return "", utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	return newAccessToken, nil
}

func generateToken(user models.UserCore, duration time.Duration, signingKey []byte) (token string, err error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.At(time.Now().Add(duration * time.Second)),
		},
		Id:   user.ID,
		Role: user.Role,
	}
	ss := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err = ss.SignedString(signingKey)
	return token, err
}

func parseToken(token string, key []byte) (*UserClaims, error) {
	data, err := jwt.ParseWithClaims(token, &UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})
	claims, ok := data.Claims.(*UserClaims)
	if err != nil {
		if claims.ExpiresAt.Unix() < time.Now().Unix() {
			return &UserClaims{}, utils.ResponseError{
				Code:    http.StatusUnauthorized,
				Message: consts.ErrTokenExpired,
			}
		}
		return &UserClaims{}, utils.ResponseError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}
	if !ok {
		return &UserClaims{}, utils.ResponseError{
			Code:    http.StatusUnauthorized,
			Message: consts.ErrNotStandardToken,
		}
	}
	return claims, nil
}
