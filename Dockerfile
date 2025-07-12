FROM alpine:3.22.0
RUN apk add --no-cache \
	curl \
	git \
	musl-dev

COPY plex2m3u /usr/bin/plex2m3u
ENTRYPOINT ["plex2m3u"]
