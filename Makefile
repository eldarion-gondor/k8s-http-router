
bindata.go: assets
	go get github.com/jteeuwen/go-bindata
	GOOS="" GOARCH="" go install github.com/jteeuwen/go-bindata/go-bindata
	go-bindata -o bindata.go -pkg="main" -prefix=assets -nocompress assets/...
	go fmt ./bindata.go

build: bindata.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 godep go build -a -installsuffix cgo -ldflags '-s' -o bin/router .
