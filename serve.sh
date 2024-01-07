#!/bin/sh

cd api
../../../tools/flatc --go *.fbs
cd ..
swag init -g internal/server/server.go
go build cmd/nsteg.go
./nsteg serve --port 8081