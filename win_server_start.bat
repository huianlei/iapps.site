@ECHO OFF
title Golang Server Sample
set base=%cd%
echo "server base dir : %base%"

rem default server mode
go run app.go 