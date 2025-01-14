package services

import (
	"awesomeProject/logger"
	"awesomeProject/models"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"net/url"
)

type MusicAPIServiceInterface interface {
	GetSongInfo(group, song string) (*models.SongDetail, error)
}

type MusicAPIService struct {
	baseURL string
	client  *http.Client
}

func NewMusicAPIService(baseURL string) *MusicAPIService {
	return &MusicAPIService{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (s *MusicAPIService) GetSongInfo(group, song string) (*models.SongDetail, error) {
	logger.Debug("Fetching song info",
		zap.String("group", group),
		zap.String("song", song))

	resp, err := s.client.Get(fmt.Sprintf(
		"%s/info?group=%s&song=%s",
		s.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(song),
	))
	if err != nil {
		logger.Info("Failed to fetch song info", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	var details models.SongDetail
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		logger.Info("Failed to decode response", zap.Error(err))
		return nil, err
	}

	logger.Debug("Successfully fetched song info")
	return &details, nil
}
