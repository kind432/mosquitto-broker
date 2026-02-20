package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/internal/services"
	"github.com/robboworld/mosquitto-broker/pkg/logger"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
)

type authHandler struct {
	loggers logger.Loggers
	auth    services.AuthService
}

func NewAuthHandler(
	loggers logger.Loggers,
	auth services.AuthService,
) *authHandler {
	return &authHandler{
		loggers: loggers,
		auth:    auth,
	}
}

func (h *authHandler) SetupAuthRoutes(router *gin.Engine) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/sign-up", h.SignUp)
		authGroup.POST("/sign-in", h.SignIn)
		authGroup.POST("/refresh-token", h.RefreshToken)
	}
}

type SignUp struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

func (h *authHandler) SignUp(c *gin.Context) {
	var input SignUp
	if err := c.ShouldBindJSON(&input); err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newUser := models.UserCore{
		Email:       input.Email,
		Password:    input.Password,
		FullName:    input.FullName,
		Role:        models.RoleUser,
		MosquittoOn: false,
	}

	err := h.auth.SignUp(newUser)
	if err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		var respErr utils.ResponseError
		if errors.As(err, &respErr) {
			c.JSON(int(respErr.Code), gin.H{"error": respErr.Message})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type SignIn struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *authHandler) SignIn(c *gin.Context) {
	var input SignIn
	if err := c.ShouldBindJSON(&input); err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.auth.SignIn(input.Email, input.Password)
	if err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		var respErr utils.ResponseError
		if errors.As(err, &respErr) {
			c.JSON(int(respErr.Code), gin.H{"error": respErr.Message})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  tokens.Access,
		"refresh_token": tokens.Refresh,
	})
}

type RefreshToken struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *authHandler) RefreshToken(c *gin.Context) {
	var input RefreshToken
	if err := c.ShouldBindJSON(&input); err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, err := h.auth.Refresh(input.RefreshToken)
	if err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		var respErr utils.ResponseError
		if errors.As(err, &respErr) {
			c.JSON(int(respErr.Code), gin.H{"error": respErr.Message})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}
