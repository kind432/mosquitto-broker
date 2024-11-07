package http

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/internal/services"
	"github.com/robboworld/mosquitto-broker/pkg/logger"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
	"net/http"
	"strconv"
)

type TopicHandler struct {
	loggers      logger.Loggers
	topicService services.TopicService
}

func (h TopicHandler) SetupTopicRoutes(router *gin.Engine) {
	topicGroup := router.Group("/topic")
	{
		topicGroup.POST("/", h.CreateTopic)
		topicGroup.PUT("/", h.UpdateTopicPermissions)
		topicGroup.GET("/:id", h.GetTopicById)
		topicGroup.GET("/", h.GetAllTopics)
		topicGroup.DELETE("/:id", h.DeleteTopic)
	}
}

type NewTopic struct {
	Name     string `json:"name"`
	CanRead  bool   `json:"can_read"`
	CanWrite bool   `json:"can_write"`
}

func (h TopicHandler) CreateTopic(c *gin.Context) {
	var input NewTopic
	if err := c.ShouldBindJSON(&input); err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.Value(consts.KeyId).(uint)
	role := c.Value(consts.KeyRole).(models.Role)
	accessRoles := []models.Role{models.RoleUser, models.RoleSuperAdmin}
	if !utils.DoesHaveRole(role, accessRoles) {
		h.loggers.Err.Printf("%s", consts.ErrAccessDenied)
		c.JSON(http.StatusForbidden, gin.H{"error": consts.ErrAccessDenied})
		return
	}

	topic := models.TopicCore{
		Name:     input.Name,
		CanRead:  input.CanRead,
		CanWrite: input.CanWrite,
		UserId:   userId,
	}
	newTopic, err := h.topicService.CreateTopic(topic, userId)
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

	topicHttp := models.TopicHTTP{}
	topicHttp.FromCore(newTopic)
	c.JSON(http.StatusOK, gin.H{"topic": topicHttp})
}

type UpdateTopicPermissions struct {
	ID       string `json:"id"`
	CanRead  bool   `json:"can_read"`
	CanWrite bool   `json:"can_write"`
}

func (h TopicHandler) UpdateTopicPermissions(c *gin.Context) {
	var input UpdateTopicPermissions
	if err := c.ShouldBindJSON(&input); err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userId := c.Value(consts.KeyId).(uint)
	role := c.Value(consts.KeyRole).(models.Role)
	accessRoles := []models.Role{models.RoleUser, models.RoleSuperAdmin}
	if !utils.DoesHaveRole(role, accessRoles) {
		h.loggers.Err.Printf("%s", consts.ErrAccessDenied)
		c.JSON(http.StatusForbidden, gin.H{"error": consts.ErrAccessDenied})
		return
	}
	atoi, err := strconv.Atoi(input.ID)
	if err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": consts.ErrAtoi})
		return
	}
	topic := models.TopicCore{
		ID:       uint(atoi),
		CanRead:  input.CanRead,
		CanWrite: input.CanWrite,
	}
	newTopic, err := h.topicService.UpdateTopicPermissions(topic, userId, role)
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

	topicHttp := models.TopicHTTP{}
	topicHttp.FromCore(newTopic)
	c.JSON(http.StatusOK, gin.H{"topic": topicHttp})
}

func (h TopicHandler) GetTopicById(c *gin.Context) {
	userId := c.Value(consts.KeyId).(uint)
	role := c.Value(consts.KeyRole).(models.Role)
	accessRoles := []models.Role{models.RoleUser, models.RoleSuperAdmin}
	if !utils.DoesHaveRole(role, accessRoles) {
		h.loggers.Err.Printf("%s", consts.ErrAccessDenied)
		c.JSON(http.StatusForbidden, gin.H{"error": consts.ErrAccessDenied})
		return
	}

	id := c.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": consts.ErrAtoi})
		return
	}
	topic, err := h.topicService.GetTopicById(uint(atoi), userId, role)
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

	topicHttp := models.TopicHTTP{}
	topicHttp.FromCore(topic)
	c.JSON(http.StatusOK, gin.H{"topic": topicHttp})
}

func (h TopicHandler) GetAllTopics(c *gin.Context) {
	userId := c.Value(consts.KeyId).(uint)
	role := c.Value(consts.KeyRole).(models.Role)
	accessRoles := []models.Role{models.RoleUser, models.RoleSuperAdmin}
	if !utils.DoesHaveRole(role, accessRoles) {
		h.loggers.Err.Printf("%s", consts.ErrAccessDenied)
		c.JSON(http.StatusForbidden, gin.H{"error": consts.ErrAccessDenied})
		return
	}

	var page, pageSize *int
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if pageSizeValue, err := strconv.Atoi(pageSizeStr); err == nil {
			pageSize = &pageSizeValue
		} else {
			h.loggers.Err.Printf("%s", pageSizeStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": consts.ErrAtoi})
			return
		}
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if pageValue, err := strconv.Atoi(pageStr); err == nil {
			page = &pageValue
		} else {
			h.loggers.Err.Printf("%s", pageStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": consts.ErrAtoi})
			return
		}
	}

	topics, countRows, err := h.topicService.GetAllTopics(page, pageSize, userId, role)
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

	topicsHttp := models.FromTopicsCore(topics)
	c.JSON(http.StatusOK, gin.H{
		"topics":     topicsHttp,
		"count_rows": countRows,
	})
}

func (h TopicHandler) DeleteTopic(c *gin.Context) {
	userId := c.Value(consts.KeyId).(uint)
	role := c.Value(consts.KeyRole).(models.Role)
	accessRoles := []models.Role{models.RoleUser, models.RoleSuperAdmin}
	if !utils.DoesHaveRole(role, accessRoles) {
		h.loggers.Err.Printf("%s", consts.ErrAccessDenied)
		c.JSON(http.StatusForbidden, gin.H{"error": consts.ErrAccessDenied})
		return
	}

	id := c.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		h.loggers.Err.Printf("%s", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": consts.ErrAtoi})
		return
	}
	err = h.topicService.DeleteTopic(uint(atoi), userId, role)
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
