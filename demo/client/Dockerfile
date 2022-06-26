# FROM golang:latest

# RUN mkdir /app
# WORKDIR /app

# RUN go mod install



# builder image
FROM amd64/golang:1.18 as builder
WORKDIR /build
ADD . /build
# RUN ls -a -l
RUN go build -o main

# generate clean, final image for end users
FROM ubuntu:18.04
WORKDIR /app
COPY --from=builder /build/main .
RUN ls -a -l
RUN chmod u+x main
# executable
ENTRYPOINT [ "/app/main" ]