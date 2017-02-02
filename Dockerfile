FROM alpine:3.2
ADD ./build/ ./bin/
ADD /etc/dockerfile.conf /etc/gandalf.conf
EXPOSE 8000
ENTRYPOINT ["/bin/gandalf-webserver"]
