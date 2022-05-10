FROM alpine:3.15.4

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

COPY terraform-checker /go/bin/terraform-checker

RUN apk add curl git openssh unzip

# Simple tfswitch
RUN curl -Ls https://github.com/terraform-tools/simple-tfswitch/releases/download/0.1.4/simple-tfswitch_0.1.4_Linux_x86_64.tar.gz | tar xzf - -C /usr/local/bin
RUN mv /usr/local/bin/simple-tfswitch /usr/local/bin/terraform

# Tflint
RUN curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | sh

# App user
RUN mkdir -p /home/appuser && chown appuser:appuser /home/appuser
USER appuser:appuser

EXPOSE 8000
WORKDIR /go

ENTRYPOINT ["/go/bin/terraform-checker"]
