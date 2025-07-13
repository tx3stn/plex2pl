FROM gcr.io/distroless/static:nonroot
COPY .schema /config
COPY plex2pl /usr/bin/plex2pl
ENTRYPOINT ["plex2pl"]
