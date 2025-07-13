FROM gcr.io/distroless/static:nonroot
COPY .schema /config
COPY plex2m3u /usr/bin/plex2m3u
ENTRYPOINT ["plex2m3u"]
