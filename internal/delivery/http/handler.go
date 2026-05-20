package http

import (
	"errors"
	"net/http"

	"booking_bot/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Handler struct {
	uc  domain.Usecase
	log *zerolog.Logger
}

func NewRouter(uc domain.Usecase, logger *zerolog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	handler := &Handler{
		uc:  uc,
		log: logger,
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.GET("/site-info", handler.GetSiteInfo)

	return router
}

func (h *Handler) GetSiteInfo(c *gin.Context) {
	rawURL := c.Query("url")
	if rawURL == "" {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Error:   "bad_request",
			Details: "query parameter url is required",
		})
		return
	}

	info, err := h.uc.CheckSiteForRegistration(c.Request.Context(), rawURL)
	if err != nil {
		h.log.Error().Err(err).Str("url", rawURL).Msg("Check site for registration request failed")

		if errors.Is(err, domain.ErrEmptyParameter) || errors.Is(err, domain.ErrURLParse) || errors.Is(err, domain.ErrInvalidParameter) {
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Error:   "bad_request",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Error:   "internal_error",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.SiteInfoResponse{
		URL:                     info.URL,
		Title:                   info.Title,
		IsRegistrationAvailable: info.IsRegistrationAvailable,
	})
}
