FROM golang:1.25-alpine AS builder
WORKDIR /app
ARG TARGETARCH
RUN apk update
RUN apk add --no-cache build-base gcc wget unzip
COPY . .
ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"
RUN go build -ldflags "-w -s" -o build/x-ui main.go
RUN ./DockerInitFiles.sh "$TARGETARCH"

FROM alpine
LABEL org.opencontainers.image.authors="alireza7@gmail.com"
ENV TZ=Asia/Tehran
WORKDIR /app

RUN apk add ca-certificates tzdata

COPY --from=builder  /app/build/ /app/
VOLUME [ "/etc/x-ui" ]
CMD [ "./x-ui" ]
