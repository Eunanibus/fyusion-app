package demo

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
)

func DemoCallbackHandler(res http.ResponseWriter, req *http.Request) {
	bodyBytes, _ := ioutil.ReadAll(req.Body)
	log.Infof("demo callback endpoint hit. response payload: %s", string(bodyBytes))
	res.WriteHeader(http.StatusOK)
}
