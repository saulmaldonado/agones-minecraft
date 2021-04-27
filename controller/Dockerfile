FROM golang:alpine3.13 AS build
WORKDIR /agones-mc-dns-controller/
RUN apk --no-cache add ca-certificates
COPY . .
RUN CGO_ENABLED=0 go build -o ./build/controller ./main.go

FROM scratch
WORKDIR /agones-minecraft-dns-controller/
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /agones-mc-dns-controller/build/controller .
ENTRYPOINT [ "./controller" ]
