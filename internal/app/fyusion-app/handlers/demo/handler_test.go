package demo

import (
	"bytes"
	"encoding/json"
	"github.com/eunanibus/fyusion-app/internal/app/fyusion-app/mapper"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDemoCallbackHandler(t *testing.T) {
	req, err := http.NewRequest(http.MethodPost, "/callback", bytes.NewBuffer(demoMediaResponse()))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(DemoCallbackHandler)
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if rr.Body.String() != "" {
		t.Error("handler returned a non-empty body")
	}
}

func demoMediaResponse() []byte {
	responseBytes, _ := json.Marshal(&mapper.MediaRequestCallbackResponse{
		Success:        true,
		OutputFilePath: "demo",
	})
	return responseBytes
}
