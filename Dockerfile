FROM golang:1.14 as builder
WORKDIR /go/src/github.com/videocoin/cloud-uploader
COPY . .
RUN make build

FROM jrottenberg/ffmpeg:4.1-ubuntu

ENV LD_LIBRARY_PATH=/usr/local/lib:/usr/local/lib64:/usr/lib:/usr/lib64:/lib:/lib64:/lib/x86_64-linux-gnu

RUN apt-get update && \
    apt-get -y --force-yes install \
        mediainfo \
        libmediainfo-dev \
        ffmpeg \
        curl

COPY --from=builder /go/src/github.com/videocoin/cloud-uploader/bin/uploader /opt/videocoin/bin/uploader

ENTRYPOINT ["/opt/videocoin/bin/uploader"]
CMD [""]
