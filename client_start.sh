#!/bin/sh

APP_HOME=$(cd "$(dirname "$0")";pwd)

echo "APP_HOME="$APP_HOME

cd $APP_HOME/../../


export GOPATH=`pwd`
echo "GOPATH="$GOPATH

cd $APP_HOME


function show_error(){
	echo -e "\e[1;41m [ERROR] $1 \e[0m"
}

function check_cmd(){
	if [ $? -ne 0 ]; then
		show_error "command $1 not found"
		exit
	fi
}

function check_env(){
	go version  >> /dev/null
	check_cmd "go"
}

check_env

go run app.go -connect localhost:9001 -count 10 -concurrent 2
