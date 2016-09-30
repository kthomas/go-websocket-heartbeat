FROM golang:onbuild
COPY key.pem key.pem
COPY cert.pem cert.pem
EXPOSE 8080
