package handler

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mairuu/mp-api/internal/features/manga/service"
	"github.com/mairuu/mp-api/internal/platform/metrics"
	httptransport "github.com/mairuu/mp-api/internal/platform/transport/http"
)

type Handler struct {
	log     *slog.Logger
	service *service.Service
}

func NewHandler(logger *slog.Logger, service *service.Service) *Handler {
	return &Handler{
		log:     logger,
		service: service,
	}
}

func (h *Handler) RegisterRoutes(router gin.IRouter) {
	mangas := router.Group("mangas")
	{
		mangas.POST("", h.CreateManga)
		mangas.GET("", h.ListMangas)
		mangas.GET(":manga_id", h.GetMangaByID)
		mangas.PUT(":manga_id", h.UpdateManga)
		mangas.DELETE(":manga_id", h.DeleteManga)
	}

	chapters := router.Group("chapters")
	{
		chapters.POST("", h.CreateChapter)
		chapters.GET("", h.ListChapters)
		chapters.GET(":chapter_id", h.GetChapterByID)
		chapters.PUT(":chapter_id", h.UpdateChapter)
		chapters.DELETE(":chapter_id", h.DeleteChapter)
	}
}

func (h *Handler) CreateManga(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	var req service.CreateMangaDTO
	if h.fail(ctx, httptransport.BindJSON(ctx, &req, h.log)) {
		return
	}

	m, err := h.service.CreateManga(ctx.Request.Context(), ur, req)
	if h.fail(ctx, err) {
		return
	}

	metrics.MangasUploadedTotal.Inc()

	httptransport.SuccessResponse(ctx, http.StatusCreated, m)
}

func (h *Handler) ListMangas(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	var q service.MangaListQuery
	if h.fail(ctx, httptransport.BindQuery(ctx, &q, h.log)) {
		return
	}

	dto, err := h.service.ListMangas(ctx.Request.Context(), ur, &q)
	if h.fail(ctx, err) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, dto)
}

func (h *Handler) GetMangaByID(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	mangaID, err := h.mangaIDFromPath(ctx)
	if h.fail(ctx, err) {
		return
	}

	dto, err := h.service.GetMangaByID(ctx.Request.Context(), ur, mangaID)
	if h.fail(ctx, err) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, dto)
}

func (h *Handler) UpdateManga(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	mangaID, err := h.mangaIDFromPath(ctx)
	if h.fail(ctx, err) {
		return
	}

	var req service.UpdateMangaDTO
	if h.fail(ctx, httptransport.BindJSON(ctx, &req, h.log)) {
		return
	}

	dto, err := h.service.UpdateManga(ctx.Request.Context(), ur, mangaID, req)
	if h.fail(ctx, err) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, dto)
}

func (h *Handler) DeleteManga(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	mangaID, err := h.mangaIDFromPath(ctx)
	if h.fail(ctx, err) {
		return
	}

	if h.fail(ctx, h.service.DeleteManga(ctx.Request.Context(), ur, mangaID)) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, nil)
}

func (h *Handler) CreateChapter(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	var req service.CreateChapterDTO
	if h.fail(ctx, httptransport.BindJSON(ctx, &req, h.log)) {
		return
	}

	chapter, err := h.service.CreateChapter(ctx.Request.Context(), ur, req)
	if h.fail(ctx, err) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusCreated, chapter)
}

func (h *Handler) ListChapters(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	var pq service.ChapterListQuery
	if h.fail(ctx, httptransport.BindQuery(ctx, &pq, h.log)) {
		return
	}

	dto, err := h.service.ListChapters(ctx.Request.Context(), ur, &pq)
	if h.fail(ctx, err) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, dto)
}

func (h *Handler) GetChapterByID(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	chapterID, err := h.chapterIDFromPath(ctx)
	if h.fail(ctx, err) {
		return
	}

	dto, err := h.service.GetChapterByID(ctx.Request.Context(), ur, chapterID)
	if h.fail(ctx, err) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, dto)
}

func (h *Handler) UpdateChapter(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	chapterID, err := h.chapterIDFromPath(ctx)
	if h.fail(ctx, err) {
		return
	}

	var req service.UpdateChapterDTO
	if h.fail(ctx, httptransport.BindJSON(ctx, &req, h.log)) {
		return
	}

	dto, err := h.service.UpdateChapter(ctx.Request.Context(), ur, chapterID, req)
	if h.fail(ctx, err) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, dto)
}

func (h *Handler) DeleteChapter(ctx *gin.Context) {
	ur := h.userRoleFromContext(ctx)

	chapterID, err := h.chapterIDFromPath(ctx)
	if h.fail(ctx, err) {
		return
	}

	if h.fail(ctx, h.service.DeleteChapter(ctx.Request.Context(), ur, chapterID)) {
		return
	}

	httptransport.SuccessResponse(ctx, http.StatusOK, nil)
}
