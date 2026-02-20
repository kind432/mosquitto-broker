package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/internal/services"
	"github.com/robboworld/mosquitto-broker/pkg/logger"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
)

type userHandler struct {
	loggers logger.Loggers
	user    services.UserService
}

func NewUserHandler(
	loggers logger.Loggers,
	user services.UserService,
) *userHandler {
	return &userHandler{
		loggers: loggers,
		user:    user,
	}
}

func (h *userHandler) SetupUserRoutes(router *gin.Engine) {
	userGroup := router.Group("/user")
	{
		userGroup.GET("/me", h.Me)
	}
}

func (h *userHandler) Me(c *gin.Context) {
	userId := c.Value(consts.KeyId).(uint)
	role := c.Value(consts.KeyRole).(models.Role)

	accessRoles := []models.Role{models.RoleUser, models.RoleSuperAdmin}
	if !utils.DoesHaveRole(role, accessRoles) {
		h.loggers.Err.Printf("%s", consts.ErrAccessDenied)
		c.JSON(http.StatusForbidden, gin.H{"error": consts.ErrAccessDenied})
		return
	}

	user, err := h.user.GetUserById(userId, userId, role)
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

	userHttp := models.UserHTTP{}
	userHttp.FromCore(user)
	c.JSON(http.StatusOK, gin.H{"user": userHttp})
}
