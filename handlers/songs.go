package handlers

import (
	"awesomeProject/logger"
	"awesomeProject/models"
	"awesomeProject/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"strconv"
	"strings"
)

type SongHandler struct {
	db       *gorm.DB
	musicAPI *services.MusicAPIService
}

func NewSongHandler(db *gorm.DB, musicAPI *services.MusicAPIService) *SongHandler {
	return &SongHandler{db: db, musicAPI: musicAPI}
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
	var songs []models.Song
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	query := h.db.Model(&models.Song{})

	if group := c.Query("group"); group != "" {
		query = query.Where("\"group\" ILIKE ?", "%"+group+"%")
	}
	if song := c.Query("song"); song != "" {
		query = query.Where("name ILIKE ?", "%"+song+"%")
	}
	if releaseDate := c.Query("releaseDate"); releaseDate != "" {
		query = query.Where("release_date = ?", releaseDate)
	}
	if link := c.Query("link"); link != "" {
		query = query.Where("link ILIKE ?", "%"+link+"%")
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * limit
	result := query.Offset(offset).Limit(limit).Find(&songs)

	if result.Error != nil {
		logger.Info("Failed to fetch songs", zap.Error(result.Error))
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

// @Summary Get song text
// @Description Get song text with pagination by verses
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param page query int false "Page number"
// @Param limit query int false "Verses per page"
// @Success 200 {object} map[string]interface{}
// @Router /songs/{id}/text [get]
func (h *SongHandler) GetText(c *gin.Context) {
	var song models.Song
	if err := h.db.First(&song, c.Param("id")).Error; err != nil {
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

// @Summary Create song
// @Description Create a new song
// @Tags songs
// @Accept json
// @Produce json
// @Param song body models.CreateSongRequest true "Song info"
// @Success 201 {object} models.Song
// @Router /songs [post]
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

	if err := h.db.Create(&song).Error; err != nil {
		logger.Info("Failed to create song", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to create song"})
		return
	}

	logger.Debug("Song created successfully", zap.String("group", song.Group), zap.String("name", song.Name))
	c.JSON(201, song)
}

// @Summary Update song
// @Description Update an existing song
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Param song body models.Song true "Updated song info"
// @Success 200 {object} models.Song
// @Router /songs/{id} [put]
func (h *SongHandler) Update(c *gin.Context) {
	var song models.Song
	if err := h.db.First(&song, c.Param("id")).Error; err != nil {
		logger.Info("Song not found", zap.Error(err))
		c.JSON(404, gin.H{"error": "Song not found"})
		return
	}

	if err := c.ShouldBindJSON(&song); err != nil {
		logger.Info("Invalid request", zap.Error(err))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Save(&song).Error; err != nil {
		logger.Info("Failed to update song", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to update song"})
		return
	}

	logger.Debug("Song updated successfully", zap.Uint("id", song.ID))
	c.JSON(200, song)
}

// @Summary Delete song
// @Description Delete a song
// @Tags songs
// @Accept json
// @Produce json
// @Param id path int true "Song ID"
// @Success 204 "No Content"
// @Router /songs/{id} [delete]
func (h *SongHandler) Delete(c *gin.Context) {
	if err := h.db.Delete(&models.Song{}, c.Param("id")).Error; err != nil {
		logger.Info("Failed to delete song", zap.Error(err))
		c.JSON(500, gin.H{"error": "Failed to delete song"})
		return
	}

	logger.Debug("Song deleted successfully", zap.String("id", c.Param("id")))
	c.Status(204)
}
