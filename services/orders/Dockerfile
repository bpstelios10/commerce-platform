FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /opt/apps

COPY bin/orders ./orders

EXPOSE 8083

ENTRYPOINT ["./orders"]
