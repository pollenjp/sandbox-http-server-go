FROM mirror.gcr.io/golang:1.19-bullseye as builder
COPY . /src
WORKDIR /src
RUN \
    go build \
    -a \
    #-ldflags="-w -s -linkmode external -extldflags '-static'" \
    -o ./build/app
RUN ldd ./build/app || true

FROM mirror.gcr.io/debian:bullseye-slim as prod
ARG APP_DIR=/usr/src/app
ARG APP_USER=appuser

COPY --from=builder /src/build/app ${APP_DIR}/app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN groupadd "${APP_USER}" \
    && useradd -g "${APP_USER}" "${APP_USER}" \
    && mkdir -p "${APP_DIR}"

RUN chown -R "${APP_USER}:${APP_USER}" "${APP_DIR}"

USER ${APP_USER}
WORKDIR ${APP_DIR}

CMD ["./app"]
