package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/internal/services"
	"github.com/robboworld/mosquitto-broker/pkg/logger"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
	"net/http"
)

type AuthHandler struct {
	loggers     logger.Loggers
	authService services.AuthService
}

func (h AuthHandler) SetupAuthRoutes(router *gin.Engine) {
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

func (h AuthHandler) SignUp(c *gin.Context) {
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

	err := h.authService.SignUp(newUser)
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

func (h AuthHandler) SignIn(c *gin.Context) {
	var input SignIn
	if err := c.ShouldBindJSON(&input); err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := h.authService.SignIn(input.Email, input.Password)
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

func (h AuthHandler) RefreshToken(c *gin.Context) {
	var input RefreshToken
	if err := c.ShouldBindJSON(&input); err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, err := h.authService.Refresh(input.RefreshToken)
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
