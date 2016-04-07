FROM alpine:3.2
ADD /webserver/webserver /bin/webserver
ADD /etc/gandalf.conf /etc/gandalf.conf
EXPOSE 8000
ENTRYPOINT ["/bin/webserver"]
