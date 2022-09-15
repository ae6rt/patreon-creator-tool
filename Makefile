DATE:=$$(date)
GOVERSION:=$$(go version)
VERSION:=$$(git rev-parse --short HEAD)

linux:
	go build -ldflags="-X 'main.version=git-hash: ${VERSION}, sdk: ${GOVERSION}, built: ${DATE}'"

mac:
	GOOS=darwin go build -ldflags="-X 'main.version=git-hash: ${VERSION}, sdk: ${GOVERSION}, built: ${DATE}'"
