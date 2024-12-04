package models

type Song struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Group       string `json:"group" binding:"required"`
	Name        string `json:"song" binding:"required"`
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type CreateSongRequest struct {
	Group string `json:"group" binding:"required,min=1"`
	Song  string `json:"song" binding:"required,min=1"`
}

type SongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}
