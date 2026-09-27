FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /opt/apps

COPY bin/products ./products

EXPOSE 8082
EXPOSE 8092

ENTRYPOINT ["./products"]
