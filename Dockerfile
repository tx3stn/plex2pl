FROM alpine:3.24 AS source
RUN mkdir /config

FROM gcr.io/distroless/static:nonroot

COPY --from=source /config /config

COPY plex2pl /usr/bin/plex2pl

ENTRYPOINT ["plex2pl"]
