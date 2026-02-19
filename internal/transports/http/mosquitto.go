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

type MosquittoHandler struct {
	loggers          logger.Loggers
	mosquittoService services.MosquittoService
}

func (h MosquittoHandler) SetupMosquittoRoutes(router *gin.Engine) {
	mosquittoGroup := router.Group("/mosquitto")
	{
		mosquittoGroup.POST("/launch", h.Launch)
	}
}

type MosquittoConfig struct {
	MosquittoOn bool `json:"mosquitto_on"`
}

func (h MosquittoHandler) Launch(c *gin.Context) {
	userId := c.Value(consts.KeyId).(uint)
	role := c.Value(consts.KeyRole).(models.Role)
	accessRoles := []models.Role{models.RoleUser, models.RoleSuperAdmin}
	if !utils.DoesHaveRole(role, accessRoles) {
		h.loggers.Err.Printf("%s", consts.ErrAccessDenied)
		c.JSON(http.StatusForbidden, gin.H{"error": consts.ErrAccessDenied})
		return
	}

	var input MosquittoConfig
	if err := c.ShouldBind(&input); err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	err := h.mosquittoService.MosquittoLaunch(userId, input.MosquittoOn)
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
