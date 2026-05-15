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

type siteInfoResponse struct {
	URL         string `json:"url"`
	Title       string `json:"title"`
	H1          string `json:"h1"`
	LinksCount  int    `json:"linksCount"`
	TextPreview string `json:"textPreview"`
}

type errorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
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
		c.JSON(http.StatusBadRequest, errorResponse{
			Error:   "bad_request",
			Details: "query parameter url is required",
		})
		return
	}

	info, err := h.uc.AnalyzeSite(c.Request.Context(), rawURL)
	if err != nil {
		h.log.Error().Err(err).Str("url", rawURL).Msg("Analyze site request failed")

		if errors.Is(err, domain.ErrEmptyParameter) || errors.Is(err, domain.ErrURLParse) || errors.Is(err, domain.ErrInvalidParameter) {
			c.JSON(http.StatusBadRequest, errorResponse{
				Error:   "bad_request",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, errorResponse{
			Error:   "internal_error",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, siteInfoResponse{
		URL:         rawURL,
		Title:       info.Title,
		H1:          info.H1,
		LinksCount:  info.LinksCount,
		TextPreview: info.TextPreview,
	})
}
