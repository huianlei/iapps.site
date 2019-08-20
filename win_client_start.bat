@ECHO OFF
title Golang Client Sample
set base=%cd%
echo "client base dir : %base%"

rem param "-connect [ip:port]" means: remote server address. required for client mode 
rem pram "-count 10000" means: total 10000 clients to be started
rem pram "-concurrent 200" means: start 200 clients  per second
  
go run app.go -connect localhost:9001 -count 100 -concurrent 10 