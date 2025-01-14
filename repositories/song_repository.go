package repositories

import (
	"awesomeProject/models"
	"gorm.io/gorm"
)

type SongRepository interface {
	List(page, limit int, filters map[string]string) ([]models.Song, int64, error)
	GetByID(id string) (*models.Song, error)
	Create(song *models.Song) error
	Update(song *models.Song) error
	Delete(id string) error
}

type SQLSongRepository struct {
	db *gorm.DB
}

func NewSQLSongRepository(db *gorm.DB) *SQLSongRepository {
	return &SQLSongRepository{db: db}
}

func (r *SQLSongRepository) List(page, limit int, filters map[string]string) ([]models.Song, int64, error) {
	var songs []models.Song
	query := r.db.Model(&models.Song{})

	if group, ok := filters["group"]; ok && group != "" {
		query = query.Where("\"group\" ILIKE ?", "%"+group+"%")
	}
	if song, ok := filters["song"]; ok && song != "" {
		query = query.Where("name ILIKE ?", "%"+song+"%")
	}
	if releaseDate, ok := filters["releaseDate"]; ok && releaseDate != "" {
		query = query.Where("release_date = ?", releaseDate)
	}
	if link, ok := filters["link"]; ok && link != "" {
		query = query.Where("link ILIKE ?", "%"+link+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Find(&songs).Error
	return songs, total, err
}

func (r *SQLSongRepository) GetByID(id string) (*models.Song, error) {
	var song models.Song
	err := r.db.First(&song, id).Error
	if err != nil {
		return nil, err
	}
	return &song, nil
}

func (r *SQLSongRepository) Create(song *models.Song) error {
	return r.db.Create(song).Error
}

func (r *SQLSongRepository) Update(song *models.Song) error {
	return r.db.Save(song).Error
}

func (r *SQLSongRepository) Delete(id string) error {
	return r.db.Delete(&models.Song{}, id).Error
}
