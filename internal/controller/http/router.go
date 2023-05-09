package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/entities"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/infrastucture/repository"
	"github.com/Imm0bilize/gunshot-telegram-notifier/internal/ucase"
)

type Handler struct {
	domain *ucase.UCase
	logger *zap.Logger
}

func NewHTTPServer(logger *zap.Logger, domain *ucase.UCase) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())

	router.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"message": "pong"}) })

	handler := &Handler{
		domain: domain,
		logger: logger,
	}

	api := router.Group("/api/v1")
	{
		api.POST("/", handler.Create)
		api.DELETE("/:id", handler.Delete)
	}

	return router
}

func (h *Handler) Create(c *gin.Context) {
	var req entities.TGAccount

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	err := h.domain.ClientUCase.Create(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, repository.ErrRecordExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "ok"})
}

func (h *Handler) Delete(c *gin.Context) {
	id := c.Param("id")

	err := h.domain.ClientUCase.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrRecordExists) {
			c.JSON(http.StatusNotFound, gin.H{"error": "the user not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
