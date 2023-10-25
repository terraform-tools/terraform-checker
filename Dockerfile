FROM alpine:3.18

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
ARG TFSWITCH_ARCH=x86_64
RUN curl -Ls https://github.com/terraform-tools/simple-tfswitch/releases/download/0.1.6/simple-tfswitch_0.1.6_Linux_${TFSWITCH_ARCH}.tar.gz | tar xzf - -C /usr/local/bin
RUN mv /usr/local/bin/simple-tfswitch /usr/local/bin/terraform

# Tflint
ARG TFLINT_ARCH=amd64
ENV TFLINT_VERSION v0.48.0
RUN wget https://github.com/terraform-linters/tflint/releases/download/${TFLINT_VERSION}/tflint_linux_${TFLINT_ARCH}.zip -O /tmp/tflint.zip && \
    unzip /tmp/tflint.zip -d /bin && \
    rm /tmp/tflint.zip

# App user
RUN mkdir -p /home/appuser && chown appuser:appuser /home/appuser
USER appuser:appuser

EXPOSE 8000
WORKDIR /go

ENTRYPOINT ["/go/bin/terraform-checker"]
