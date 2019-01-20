package mapper

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	appStoragePath          = "./tmp"
	imageStorageFolderName  = "images"
	outputStorageFolderName = "output"
)

type APIError struct {
	ID        string    `json:"id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type MediaRequestError struct {
	statusCode int
	error      error
}

type MediaRequestCallbackResponse struct {
	Success        bool   `json:"operation_successful"`
	OutputFilePath string `json:"output_path,omitempty"`
}

func NewMediaRequestCallbackResponse(mr *MediaRequest) *MediaRequestCallbackResponse {
	return &MediaRequestCallbackResponse{
		Success:        true,
		OutputFilePath: mr.OutputStoragePath(),
	}
}

func (cr *MediaRequestCallbackResponse) ToggleSuccessFailed() {
	cr.Success = false
	cr.OutputFilePath = ""
}

func NewMediaRequestError(statusCode int, err error) *MediaRequestError {
	return &MediaRequestError{statusCode, err}
}

func (e *MediaRequestError) GetError() error {
	return e.error
}

func (e *MediaRequestError) GetStatusCode() int {
	return e.statusCode
}

type MediaRequest struct {
	ID                string
	rootStoragePath   string
	videoStoragePath  string
	imageStoragePath  string
	outputStoragePath string
	CallbackURL       string
	fileEnding        string
}

// NewMediaRequest represents an incoming media conversion request. It bundles ID generation and path creation
// logic, as well as getter functions into a centralised struct
func NewMediaRequest(callbackURL string, fileBytes []byte) (*MediaRequest, *MediaRequestError) {
	fileType := http.DetectContentType(fileBytes)
	if fileType != "video/mp4" {
		return nil, NewMediaRequestError(http.StatusUnsupportedMediaType, errors.New("invalid file type. File must be of type mp4"))
	}

	fileEndings, err := mime.ExtensionsByType(fileType)
	if err != nil {
		return nil, NewMediaRequestError(http.StatusUnsupportedMediaType, errors.New("cannot read file type"))
	}

	fileID := uuid.Must(uuid.NewV4()).String()

	mReq := &MediaRequest{
		ID:              fileID,
		CallbackURL:     callbackURL,
		rootStoragePath: fmt.Sprintf("%s/%s/", appStoragePath, fileID),
		fileEnding:      fileEndings[0],
	}

	if err := os.MkdirAll(mReq.ImageStoragePath(), os.ModePerm); err != nil {
		return nil, NewMediaRequestError(http.StatusInternalServerError, err)
	}

	if err := os.MkdirAll(mReq.OutputStoragePath(), os.ModePerm); err != nil {
		return nil, NewMediaRequestError(http.StatusInternalServerError, err)
	}

	newFile, err := os.Create(mReq.VideoStoragePath())
	if err != nil {
		return nil, NewMediaRequestError(http.StatusInternalServerError, err)
	}
	defer newFile.Close()

	if _, err := newFile.Write(fileBytes); err != nil {
		return nil, NewMediaRequestError(http.StatusInternalServerError, errors.New("cannot write file to local storage"))
	}
	return mReq, nil
}

func (m *MediaRequest) ImageStoragePath() string {
	return filepath.Join(m.rootStoragePath, imageStorageFolderName)
}

func (m *MediaRequest) OutputStoragePath() string {
	return filepath.Join(m.rootStoragePath, outputStorageFolderName)
}

func (m *MediaRequest) VideoStoragePath() string {
	return filepath.Join(m.rootStoragePath, m.ID+m.fileEnding)
}

func (m *MediaRequest) GenerateFileOutputName(file os.FileInfo) string {
	return filepath.Join(m.OutputStoragePath(), file.Name())
}

func (m *MediaRequest) GetImageFilePath(file os.FileInfo) string {
	return filepath.Join(m.ImageStoragePath(), file.Name())
}
