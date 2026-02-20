package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/robboworld/mosquitto-broker/internal/consts"
	"github.com/robboworld/mosquitto-broker/internal/models"
	"github.com/robboworld/mosquitto-broker/internal/services"
	"github.com/robboworld/mosquitto-broker/pkg/logger"
	"github.com/robboworld/mosquitto-broker/pkg/utils"
)

type topicHandler struct {
	loggers logger.Loggers
	topic   services.TopicService
}

func NewTopicHandler(
	loggers logger.Loggers,
	topic services.TopicService,
) *topicHandler {
	return &topicHandler{
		loggers: loggers,
		topic:   topic,
	}
}

func (h *topicHandler) SetupTopicRoutes(router *gin.Engine) {
	topicGroup := router.Group("/topic")
	{
		topicGroup.POST("/", h.Create)
		topicGroup.GET("/:id", h.GetById)
		topicGroup.GET("/", h.GetAll)
		topicGroup.PUT("/", h.UpdatePermissions)
		topicGroup.DELETE("/:id", h.Delete)
	}
}

type NewTopic struct {
	Name     string `json:"name"`
	CanRead  bool   `json:"can_read"`
	CanWrite bool   `json:"can_write"`
}

func (h *topicHandler) Create(c *gin.Context) {
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

	newTopic, err := h.topic.Create(topic, userId)
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

func (h *topicHandler) GetById(c *gin.Context) {
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

	topic, err := h.topic.GetById(uint(atoi), userId, role)
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

func (h *topicHandler) GetAll(c *gin.Context) {
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

	topics, countRows, err := h.topic.GetAll(page, pageSize, userId, role)
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

type UpdateTopicPermissions struct {
	ID       string `json:"id"`
	CanRead  bool   `json:"can_read"`
	CanWrite bool   `json:"can_write"`
}

func (h *topicHandler) UpdatePermissions(c *gin.Context) {
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

	updatedTopic, err := h.topic.UpdatePermissions(topic, userId, role)
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
	topicHttp.FromCore(updatedTopic)
	c.JSON(http.StatusOK, gin.H{"topic": topicHttp})
}

func (h *topicHandler) Delete(c *gin.Context) {
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

	err = h.topic.Delete(uint(atoi), userId, role)
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
