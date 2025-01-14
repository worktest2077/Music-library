package handlers

import (
	"awesomeProject/logger"
	"awesomeProject/models"
	"awesomeProject/repositories"
	"awesomeProject/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type SongHandler struct {
	songRepo repositories.SongRepository
	musicAPI services.MusicAPIServiceInterface // изменили тип на интерфейс
}

func NewSongHandler(repo repositories.SongRepository, api services.MusicAPIServiceInterface) *SongHandler {
	return &SongHandler{songRepo: repo, musicAPI: api}
}

// @Summary Get songs list
// @Description Get list of songs with pagination and filtering
// @Tags songs
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Param group query string false "Filter by group"
// @Param song query string false "Filter by song name"
// @Param releaseDate query string false "Filter by release date"
// @Success 200 {object} models.Song
// @Router /songs [get]
func (h *SongHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	filters := map[string]string{
		"group":       c.Query("group"),
		"song":        c.Query("song"),
		"releaseDate": c.Query("releaseDate"),
		"link":        c.Query("link"),
	}

	songs, total, err := h.songRepo.List(page, limit, filters)
	if err != nil {
		logger.Info("Failed to fetch songs", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to fetch songs"})
		return
	}

	logger.Debug("Successfully fetched songs",
		zap.Int("count", len(songs)),
		zap.Int("page", page),
		zap.Int("limit", limit))

	c.JSON(200, gin.H{
		"total": total,
		"items": songs,
	})
}

func (h *SongHandler) GetText(c *gin.Context) {
	song, err := h.songRepo.GetByID(c.Param("id"))
	if err != nil {
		logger.Info("Song not found", zap.Error(err))
		c.JSON(404, gin.H{"error": "Song not found"})
		return
	}

	verses := strings.Split(song.Text, "\n\n")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "1"))

	start := (page - 1) * limit
	end := start + limit
	if end > len(verses) {
		end = len(verses)
	}

	logger.Debug("Fetching song text",
		zap.Int("songId", int(song.ID)),
		zap.Int("page", page),
		zap.Int("limit", limit))

	c.JSON(200, gin.H{
		"total":  len(verses),
		"verses": verses[start:end],
	})
}

func (h *SongHandler) Create(c *gin.Context) {
	var req models.CreateSongRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Info("Invalid request", zap.Error(err))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	details, err := h.musicAPI.GetSongInfo(req.Group, req.Song)
	if err != nil {
		logger.Info("Failed to fetch song details", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to fetch song details"})
		return
	}

	song := models.Song{
		Group:       req.Group,
		Name:        req.Song,
		ReleaseDate: details.ReleaseDate,
		Text:        details.Text,
		Link:        details.Link,
	}

	if err := h.songRepo.Create(&song); err != nil {
		logger.Info("Failed to create song", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to create song"})
		return
	}

	logger.Debug("Song created successfully", zap.String("group", song.Group), zap.String("name", song.Name))
	c.JSON(201, song)
}

func (h *SongHandler) Update(c *gin.Context) {
	song, err := h.songRepo.GetByID(c.Param("id"))
	if err != nil {
		logger.Info("Song not found", zap.Error(err))
		c.JSON(404, gin.H{"error": "Song not found"})
		return
	}

	if err := c.ShouldBindJSON(song); err != nil {
		logger.Info("Invalid request", zap.Error(err))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.songRepo.Update(song); err != nil {
		logger.Info("Failed to update song", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to update song"})
		return
	}

	logger.Debug("Song updated successfully", zap.Uint("id", song.ID))
	c.JSON(200, song)
}

func (h *SongHandler) Delete(c *gin.Context) {
	if err := h.songRepo.Delete(c.Param("id")); err != nil {
		logger.Info("Failed to delete song", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to delete song"})
		return
	}

	logger.Debug("Song deleted successfully", zap.String("id", c.Param("id")))
	c.Status(204)
}
