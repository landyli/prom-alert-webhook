export GOPROXY=https://goproxy.cn
export GOOS=linux
export GOARCH=386
set GO11MODULE=on
set GO111MODULE=on
go mod init prom-alert-webhook
go mod vendor
go build -o prometheus-alert-webhook