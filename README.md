# Fyusion Image Processing Assignment

A sample app that accepts a POST request via the default endpoint, accepting a media file upload as form input, in addition to a callback URL as a URL parameter.

#### Example Curl Request

```bash
curl -X PUT -F "file=@fyuse_h264_3.mp4" http://localhost:8888/\?callback\=http://localhost:8888/callback
```

Note: The demo callback endpoint `localhost:8888/callback` will return a default 200 response.

 
 #### Run it
 
 Execute the server by running the command `make run` from the project root. Alternatively, you can execute the command `go run cmd/fyusion-converter/main.go`
 
 #### Test it
 
 Execute tests via `make test` or `go test -v ./...`