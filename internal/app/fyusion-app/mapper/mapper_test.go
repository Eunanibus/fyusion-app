package mapper_test

import (
	"github.com/eunanibus/fyusion-app/internal/app/fyusion-app/mapper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

const (
	validCallbackURL = "http://localhost:8888/callback=http://localhost:8888"
)

func TestSuccessfullyCreatedMediaRequest(t *testing.T) {
	fileBytes, err := ioutil.ReadFile("demo.mp4")
	assert.Nil(t, err)

	mediaReq, mediaReqErr := mapper.NewMediaRequest(validCallbackURL, fileBytes)
	assert.Nil(t, mediaReqErr)
	assert.NotNil(t, mediaReq.ID)
	assert.NotNil(t, mediaReq.VideoStoragePath())
	assert.NotNil(t, mediaReq.ImageStoragePath())
	assert.NotNil(t, mediaReq.OutputStoragePath())
	assert.Equal(t, validCallbackURL, mediaReq.CallbackURL)
}

func TestInvalidMediaTypeFileProvided(t *testing.T) {
	_, mediaReqErr := mapper.NewMediaRequest(validCallbackURL, []byte("invalid file bytes"))
	assert.NotNil(t, mediaReqErr)
	assert.Equal(t, http.StatusUnsupportedMediaType, mediaReqErr.GetStatusCode())
	assert.Equal(t, "invalid file type. File must be of type mp4", mediaReqErr.GetError().Error())
}
