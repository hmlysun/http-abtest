build:
	@echo building on Mac OS
	@go build -ldflags "-s -w" -o abtest_mac ab.go config.go hashset.go util.go logger.go
build-mac-so:
	@echo building libzd.so on Mac OS
	@go build -ldflags "-s -w" -buildmode=c-shared -o libzd/libzd.so libzd.go util.go
build-linux:
	@echo building on Linux CentOS
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o abtest_linux ab.go config.go hashset.go util.go logger.go
build-linux-so:
	@echo building libzd.so on Linux CentOS
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -buildmode=c-shared -o libzd/libzd.so libzd.go util.go
clean:
	@echo clean all
	@rm -f abtest_mac abtest_linux libzd/libzd.so  libzd/libzd.h
