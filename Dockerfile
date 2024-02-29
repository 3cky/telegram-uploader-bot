FROM --platform=$BUILDPLATFORM golang:1.22-alpine AS builder

ARG VERSION
ARG TIMESTAMP

WORKDIR /src
COPY ${PWD} /src

ARG TARGETOS TARGETARCH
ENV GOOS $TARGETOS
ENV GOARCH $TARGETARCH
RUN go build -a -ldflags "-s -w -X github.com/3cky/telegram-uploader-bot/build.Version=$VERSION -X github.com/3cky/telegram-uploader-bot/build.Timestamp=$TIMESTAMP" -o bin/telegram-uploader-bot

FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /src/bin/telegram-uploader-bot /
USER 65534:65534
ENTRYPOINT ["/telegram-uploader-bot"]
