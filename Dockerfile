FROM golang:1.14 as builder
WORKDIR /go/src/github.com/videocoin/cloud-uploader
COPY . .
RUN make build

FROM jrottenberg/ffmpeg:4.1-ubuntu
RUN apt-get update && \
    apt-get -y --force-yes install \
        mediainfo \
        libmediainfo-dev \
        ffmpeg \
        curl

COPY --from=builder /go/src/github.com/videocoin/cloud-uploader/bin/uploader /opt/videocoin/bin/uploader

ENTRYPOINT ["/opt/videocoin/bin/uploader"]
CMD [""]
