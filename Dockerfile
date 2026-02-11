FROM docker.io/alpine AS builder

LABEL maintainer="vladimir@hodakov.me"

COPY . /src

RUN apk add --no-cache go

RUN cd /src && go build ./cmd/faketunes

FROM docker.io/alpine

RUN apk add --no-cache fuse3 ffmpeg

COPY --from=builder /src/faketunes /bin/faketunes

ENTRYPOINT [ "/bin/faketunes" ]
