package media_test

import (
	"encoding/json"
	"github.com/eunanibus/fyusion-app/internal/app/fyusion-app/handlers/media"
	"github.com/eunanibus/fyusion-app/internal/app/fyusion-app/mapper"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInvalidURLRejected(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/callback", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(media.HandleMediaUpload)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	errResp := mapper.APIError{}
	err = json.Unmarshal(rr.Body.Bytes(), &errResp)
	if err != nil {
		t.Errorf("unable to marshal expected response")
	}

	assert.Equal(t, errResp.Message, "Invalid demo URL provided")
}
