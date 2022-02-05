#! /bin/bash
SERVER_NAME=AWS_SSH_NAME
swag init
env GOOS=linux GOARCH=amd64 go build
scp wos-core-go $SERVER_NAME:
ssh $SERVER_NAME 'sudo ./restart-v1wos.sh; sudo ./restart-v1wosstage.sh'
