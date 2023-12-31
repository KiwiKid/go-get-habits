FROM golang:1.20.4 AS build
WORKDIR /go/src/app
COPY . .
ENV CGO_ENABLED=1 GOOS=linux GOPROXY=direct
RUN go build -v -o app .
RUN chmod +x app

FROM golang:1.20.4
COPY --from=build /go/src/app/app /go/bin/app
CMD ["/go/bin/app"]
EXPOSE 8122
VOLUME ["/db"]
ENV MQTT_URL=localhost
ENV DB_PATH=/db/habits.db