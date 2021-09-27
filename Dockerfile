FROM golang:1.17-alpine AS builder

# Create appuser.
ENV USER=appuser
ENV UID=10001
# See https://stackoverflow.com/a/55757473/12429735RUN
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/home/appuser" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR $GOPATH/src/terraform-checker
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go mod verify

COPY main.go .
COPY pkg pkg


ARG GIT_COMMIT
ARG GIT_TAG

RUN ls
RUN ls; CGO_ENABLED=0 GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o /go/bin/terraform-checker -ldflags="-extldflags '-static' -w -s -X main.buildTag=${GIT_TAG} -X main.buildRevision=${GIT_COMMIT}" -gcflags="-trimpath=${GOPATH}/src"

# Final image
FROM alpine:3.14.2

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /go/bin/terraform-checker /go/bin/terraform-checker

RUN apk add curl git openssh
RUN curl -Ls https://github.com/terraform-tools/simple-tfswitch/releases/download/0.1.1/simple-tfswitch_0.1.1_Linux_x86_64.tar.gz | tar xzf - -C /usr/local/bin
RUN mv /usr/local/bin/simple-tfswitch /usr/local/bin/terraform
RUN mkdir -p /home/appuser && chown appuser:appuser /home/appuser
USER appuser:appuser

EXPOSE 8000
WORKDIR /go

ENTRYPOINT ["/go/bin/terraform-checker"]
