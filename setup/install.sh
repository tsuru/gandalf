#!/bin/bash -e

# copy gandalf command to a place in path
go build -o gandalf bin/gandalf.go
sudo mv gandalf /usr/local/bin/

# copy default config file
sudo cp etc/gandalf.conf /etc/

# starts gandalf api web server
go build -o apiwebserver webserver/main.go
./apiwebserver > /dev/null 2>&1 &
