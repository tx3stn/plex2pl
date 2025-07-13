# TODO: check if scracth container would be fine here.
FROM alpine:3.22.0
COPY plex2m3u /usr/bin/plex2m3u
ENTRYPOINT ["plex2m3u"]
