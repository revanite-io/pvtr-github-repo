FROM alpine:3.22 AS core
RUN apk add --no-cache wget tar unzip

WORKDIR /app
ARG VERSION=0.9.1
ARG PLATFORM=Linux_x86_64  # Change this based on your target system

RUN wget https://github.com/privateerproj/privateer/releases/download/v${VERSION}/privateer_${PLATFORM}.tar.gz
RUN tar -xzf privateer_${PLATFORM}.tar.gz

FROM golang:1.25.1-alpine3.22 AS plugin
RUN apk add --no-cache make git
WORKDIR /plugin
COPY . .
RUN make binary

FROM golang:1.25.1-alpine3.22
RUN addgroup -g 1001 -S appgroup && adduser -u 1001 -S appuser -G appgroup

RUN mkdir -p /.privateer/bin && chown -R appuser:appgroup /.privateer
WORKDIR /.privateer/bin
USER appuser

COPY --from=core /app/privateer .
COPY --from=plugin /plugin/github-repo .
COPY --from=plugin /plugin/container-entrypoint.sh .

# The config file must be provided at run time.
# example: docker run -v /path/to/config.yml:/.privateer/config.yml privateer-image
CMD ["./container-entrypoint.sh"]
