package handlers

import (
	"awesomeProject/logger"
	"awesomeProject/models"
	"awesomeProject/repositories"
	"awesomeProject/services"
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockSongRepository struct {
	mock.Mock
}

func (m *MockSongRepository) List(page int, limit int, filters map[string]string) ([]models.Song, int64, error) {
	args := m.Called(page, limit, filters)
	return args.Get(0).([]models.Song), args.Get(1).(int64), args.Error(2)
}

func (m *MockSongRepository) GetByID(id string) (*models.Song, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Song), args.Error(1)
}

func (m *MockSongRepository) Create(song *models.Song) error {
	args := m.Called(song)
	return args.Error(0)
}

func (m *MockSongRepository) Update(song *models.Song) error {
	args := m.Called(song)
	return args.Error(0)
}

func (m *MockSongRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type MockMusicAPIService struct {
	mock.Mock
}

func (m *MockMusicAPIService) GetSongInfo(group, song string) (*models.SongDetail, error) {
	args := m.Called(group, song)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SongDetail), args.Error(1)
}

var _ repositories.SongRepository = (*MockSongRepository)(nil)
var _ services.MusicAPIServiceInterface = (*MockMusicAPIService)(nil)

func setupTest() (*MockSongRepository, *MockMusicAPIService, *gin.Engine) {
	logger.Init()
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockSongRepository)
	mockAPI := new(MockMusicAPIService)
	handler := NewSongHandler(mockRepo, mockAPI)

	r := gin.New()
	r.GET("/api/v1/song", handler.List)
	r.GET("/api/v1/song/:id/text", handler.GetText)
	r.POST("/api/v1/song", handler.Create)
	r.PUT("/api/v1/song/:id", handler.Update)
	r.DELETE("/api/v1/song/:id", handler.Delete)

	return mockRepo, mockAPI, r
}

func TestSongHandler_List(t *testing.T) {
	mockRepo, _, r := setupTest()

	testSongs := []models.Song{
		{
			ID:    1,
			Group: "Muse",
			Name:  "Supermassive Black Hole",
		},
		{
			ID:    2,
			Group: "Queen",
			Name:  "Bohemian Rhapsody",
		},
	}

	t.Run("Successfully get all songs", func(t *testing.T) {
		mockRepo.On("List", 1, 10, map[string]string{
			"group":       "",
			"song":        "",
			"releaseDate": "",
			"link":        "",
		}).Return(testSongs, int64(2), nil).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/song", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(2), response["total"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("Filter by group", func(t *testing.T) {
		filteredSongs := []models.Song{testSongs[0]}
		mockRepo.On("List", 1, 10, map[string]string{
			"group":       "Muse",
			"song":        "",
			"releaseDate": "",
			"link":        "",
		}).Return(filteredSongs, int64(1), nil).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/song?group=Muse", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(1), response["total"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("Database error", func(t *testing.T) {
		mockRepo.On("List", 1, 10, mock.Anything).
			Return([]models.Song{}, int64(0), errors.New("database error")).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/song", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestSongHandler_Create(t *testing.T) {
	mockRepo, mockAPI, r := setupTest()

	t.Run("Successfully create song", func(t *testing.T) {
		songDetail := &models.SongDetail{
			ReleaseDate: "2006-07-16",
			Text:        "Some lyrics",
			Link:        "https://example.com",
		}

		createReq := models.CreateSongRequest{
			Group: "Muse",
			Song:  "Supermassive Black Hole",
		}

		mockAPI.On("GetSongInfo", createReq.Group, createReq.Song).
			Return(songDetail, nil).Once()

		expectedSong := &models.Song{
			Group:       createReq.Group,
			Name:        createReq.Song,
			ReleaseDate: songDetail.ReleaseDate,
			Text:        songDetail.Text,
			Link:        songDetail.Link,
		}

		mockRepo.On("Create", mock.MatchedBy(func(s *models.Song) bool {
			return s.Group == expectedSong.Group && s.Name == expectedSong.Name
		})).Return(nil).Once()

		body, _ := json.Marshal(createReq)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/song", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockAPI.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/song", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestSongHandler_GetText(t *testing.T) {
	mockRepo, _, r := setupTest()

	t.Run("Successfully get song text", func(t *testing.T) {
		song := &models.Song{
			ID:    1,
			Text:  "Verse 1\n\nVerse 2\n\nVerse 3",
			Group: "Muse",
			Name:  "Supermassive Black Hole",
		}

		mockRepo.On("GetByID", "1").Return(song, nil).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/song/1/text", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, float64(3), response["total"])

		mockRepo.AssertExpectations(t)
	})

	t.Run("Song not found", func(t *testing.T) {
		mockRepo.On("GetByID", "999").
			Return(nil, errors.New("record not found")).Once()

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/song/999/text", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})
}
