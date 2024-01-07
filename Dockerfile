FROM alpine:latest
ARG TARGETARCH
ARG APP

EXPOSE 1883 8883 8083 8084 8421 8422

WORKDIR /app
COPY ./app/${APP}-${TARGETARCH} /app/${APP}
COPY ./config/gomq.toml /etc/gomq/

ENV cmd="/app/$APP"
ENTRYPOINT ${cmd} start
