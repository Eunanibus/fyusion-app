package media

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/eunanibus/fyusion-app/internal/app/fyusion-app/mapper"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	log "github.com/sirupsen/logrus"
	"image"
	"image/jpeg"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const (
	maxResponseRetries = 10
	maxGoRoutines      = 5
	maxUploadFileSize  = 1 * 1000 * 1024 // 1 GB Maximum file size
	outputJpegQuality  = 80
)

func HandleMediaUpload(res http.ResponseWriter, req *http.Request) {
	log.Debugf("new request received from: %s", req.RemoteAddr)
	callbackURL := mux.Vars(req)["callback"]
	if _, err := url.ParseRequestURI(callbackURL); err != nil {
		handleError(http.StatusBadRequest, "Invalid demo URL provided", err, res)
		return
	}

	req.Body = http.MaxBytesReader(res, req.Body, maxUploadFileSize)
	if err := req.ParseMultipartForm(maxUploadFileSize); err != nil {
		handleError(http.StatusRequestEntityTooLarge, err.Error(), err, res)
		return
	}

	file, _, err := req.FormFile("file")
	if err != nil {
		handleError(http.StatusBadRequest, err.Error(), err, res)
		return
	}
	defer file.Close()

	fileBytes, err := ioutil.ReadAll(file)

	mediaRequest, mediaReqError := mapper.NewMediaRequest(callbackURL, fileBytes)
	if mediaReqError != nil {
		handleError(mediaReqError.GetStatusCode(), mediaReqError.GetError().Error(), mediaReqError.GetError(), res)
	} else {
		log.Debugf("new video request successfully written to %s", mediaRequest.VideoStoragePath())
		go processVideo(mediaRequest)
		res.WriteHeader(http.StatusOK)
	}
}

func processVideo(mediaReq *mapper.MediaRequest) {
	requestResponse := mapper.NewMediaRequestCallbackResponse(mediaReq)
	if err := extractFrames(mediaReq); err != nil {
		log.Errorf("error occurred when attempting to extract frames from video id: %s ", mediaReq.ID)
		requestResponse.ToggleSuccessFailed()
	}
	if err := convertFrames(mediaReq); err != nil {
		log.Errorf("error occurred when attempting to convert frames for video id: %s", mediaReq.ID)
		requestResponse.ToggleSuccessFailed()
	}
	if err := handleCallbackResponse(mediaReq, requestResponse); err != nil {
		log.Errorf("response to demo URL failed for video ID: %s - %s", mediaReq.ID, err)
	}
}

func extractFrames(mediaReq *mapper.MediaRequest) error {
	log.Debugf("attempting to extract frames from video id: %s", mediaReq.ID)

	// We use ffmpeg cli to extract the frames as it is more efficient, and a simpler way of performing the task
	fullCommand := fmt.Sprintf(
		"ffmpeg -i %s -f image2 -start_number 0 -q:v 0 %s/frames_%%d.jpg",
		mediaReq.VideoStoragePath(),
		mediaReq.ImageStoragePath(),
	)
	commandSegments := strings.Split(fullCommand, " ")
	command := commandSegments[0]
	args := commandSegments[1:]
	cmd := exec.Command(command, args...)

	_, err := cmd.CombinedOutput()
	return err
}

func convertFrames(mediaReq *mapper.MediaRequest) error {
	files, err := ioutil.ReadDir(mediaReq.ImageStoragePath())
	if err != nil {
		log.Errorf("failed to retrieve images for video id: %s", mediaReq.ID)
		return err
	}

	fileCount := len(files)
	log.Debugf("attempting to convert %d frames for video id: %s", fileCount, mediaReq.ID)

	var wg sync.WaitGroup
	wg.Add(fileCount)
	guard := make(chan struct{}, maxGoRoutines)

	for i, file := range files {
		guard <- struct{}{}
		go func(fileNo int, file os.FileInfo) {
			log.Debugf("attempting to convert frame %d of %d for video id: %s", fileNo, fileCount, mediaReq.ID)
			if err := greyscaleImage(&wg, file, mediaReq); err != nil {
				log.Errorf("attempt to convert frame %d of %d for video id: %s failed", fileNo, fileCount, mediaReq.ID)
			}
			<-guard
		}(i+1, file)
	}

	wg.Wait()
	log.Debugf("conversion of frames for video id: %s complete", mediaReq.ID)
	return nil
}

func greyscaleImage(wg *sync.WaitGroup, file os.FileInfo, mediaReq *mapper.MediaRequest) error {
	defer wg.Done()
	filename := mediaReq.GetImageFilePath(file)

	imageFile, err := os.Open(filename)
	if err != nil {
		return err
	}

	src, _, err := image.Decode(imageFile)
	if err != nil {
		return err
	}

	imgWidth := src.Bounds().Max.X
	imgHeight := src.Bounds().Max.Y

	// Create a new grayscale image
	gray := image.NewGray(image.Rectangle{Max: image.Point{X: imgWidth, Y: imgHeight}})
	for x := 0; x < imgWidth; x++ {
		for y := 0; y < imgHeight; y++ {
			gray.Set(x, y, src.At(x, y))
		}
	}

	imageFile.Close()
	outfile, _ := os.Create(mediaReq.GenerateFileOutputName(file))

	return jpeg.Encode(outfile, gray, &jpeg.Options{Quality: outputJpegQuality})
}

func handleCallbackResponse(mediaReq *mapper.MediaRequest, callbackResp *mapper.MediaRequestCallbackResponse) error {
	log.Debugf("attempting callback for video ID: %s", mediaReq.ID)
	responseBody, err := json.Marshal(callbackResp)
	if err != nil {
		log.Errorf("failed to marshal response for video ID: %s", mediaReq.ID)
		return err
	}
	req, err := http.NewRequest(
		http.MethodPost,
		mediaReq.CallbackURL,
		bytes.NewBuffer(responseBody),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	callbackSuccess := false
	for i := 1; i <= maxResponseRetries; i++ {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			log.Debugf("callback for video ID: %s successful", mediaReq.ID)
			callbackSuccess = true
			break
		}
		log.Errorf("callback attempt #%d for video ID: %s failed", i, mediaReq.ID)

		// We increase the demo response attempts exponentially
		time.Sleep(time.Duration(i) * 10 * time.Second)
	}

	if !callbackSuccess {
		return errors.New(fmt.Sprintf("attempts respond to callback URL: %s failed after %d tries", mediaReq.CallbackURL, maxResponseRetries))
	}
	return nil
}

func handleError(statusCode int, errMsg string, err error, res http.ResponseWriter) {
	jsonBytes, err := json.Marshal(mapper.APIError{
		ID:        uuid.Must(uuid.NewV4()).String(),
		Message:   errMsg,
		CreatedAt: time.Now(),
	})

	if err != nil {
		log.WithError(err)
		http.Error(res, "failed to marshal api error to json bytes", http.StatusInternalServerError)
		return
	}

	http.Error(res, string(jsonBytes), statusCode)
}
