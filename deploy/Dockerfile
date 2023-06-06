FROM gcr.io/distroless/static

USER nonroot:nonroot

COPY --chown=nonroot:nonroot ./release /swpc
WORKDIR /swpc

EXPOSE 5000

ENTRYPOINT ["./swpc-server"]