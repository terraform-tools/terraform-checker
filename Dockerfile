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
RUN curl -Ls https://github.com/terraform-tools/simple-tfswitch/releases/download/0.1.5/simple-tfswitch_0.1.5_Linux_x86_64.tar.gz | tar xzf - -C /usr/local/bin
RUN mv /usr/local/bin/simple-tfswitch /usr/local/bin/terraform

# Tflint
ENV TFLINT_VERSION v0.37.0
RUN wget https://github.com/terraform-linters/tflint/releases/download/${TFLINT_VERSION}/tflint_linux_amd64.zip -O /tmp/tflint.zip && \
    unzip /tmp/tflint.zip -d /bin && \
    rm /tmp/tflint.zip

# App user
RUN mkdir -p /home/appuser && chown appuser:appuser /home/appuser
USER appuser:appuser

EXPOSE 8000
WORKDIR /go

ENTRYPOINT ["/go/bin/terraform-checker"]
