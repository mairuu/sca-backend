package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mairuu/mp-api/internal/features/user/service"
	"github.com/mairuu/mp-api/internal/platform/metrics"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
	"github.com/mairuu/mp-api/internal/platform/transport/http/middleware"
)

type UserHandler struct {
	log     *slog.Logger
	service *service.Service
}

func NewUserHandler(logger *slog.Logger, service *service.Service) *UserHandler {
	return &UserHandler{log: logger, service: service}
}

func (h *UserHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/register", h.Register)
	router.POST("/login", h.Login)
	router.POST("/refresh", h.Refresh)
	router.POST("/logout", h.Logout)

	router.GET("/me", middleware.RequiredAuth(), h.GetMe)
}

func (h *UserHandler) Register(ctx *gin.Context) {
	var req service.RegisterDTO
	if h.fail(ctx, httptransport.BindJSON(ctx, &req, h.log)) {
		return
	}

	user, err := h.service.Register(ctx.Request.Context(), req)
	if h.fail(ctx, err) {
		return
	}

	metrics.RegistrationsTotal.Inc()

	httptransport.SuccessResponse(ctx, http.StatusCreated, user)
}

func (h *UserHandler) Login(ctx *gin.Context) {
	var req service.LoginDTO
	if h.fail(ctx, httptransport.BindJSON(ctx, &req, h.log)) {
		return
	}

	response, err := h.service.Login(ctx.Request.Context(), req)
	if h.fail(ctx, err) {
		return
	}

	metrics.LoginsTotal.Inc()

	httptransport.SuccessResponse(ctx, http.StatusOK, response)
}

func (h *UserHandler) GetMe(ctx *gin.Context) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		httptransport.ErrorResponse(ctx, http.StatusUnauthorized, "user not authenticated")
		return
	}

	user, err := h.service.GetUserByID(ctx.Request.Context(), userID)
	if h.fail(ctx, err) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, user)
}

func (h *UserHandler) Refresh(ctx *gin.Context) {
	var req service.RefreshTokenDTO
	if h.fail(ctx, httptransport.BindJSON(ctx, &req, h.log)) {
		return
	}

	response, err := h.service.RefreshToken(ctx.Request.Context(), req)
	if h.fail(ctx, err) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, response)
}

func (h *UserHandler) Logout(ctx *gin.Context) {
	var req service.RefreshTokenDTO
	if h.fail(ctx, httptransport.BindJSON(ctx, &req, h.log)) {
		return
	}

	if h.fail(ctx, h.service.Logout(ctx.Request.Context(), req)) {
		return
	}

	// response body need to align with another enpoints for simplicity in client handling
	httptransport.SuccessResponse(ctx, http.StatusOK, nil)
}
