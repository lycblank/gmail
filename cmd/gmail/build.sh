CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -o gmail_mac
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o gmail_linux
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -o gmail_windows.exe
