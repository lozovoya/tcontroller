FROM golang:1.17-alpine AS build
ADD . /ticketsystemcontroller
ENV CGO_ENABLED=0
WORKDIR /ticketsystemcontroller
RUN go build -o ticketsystemcontroller.bin ./cmd/ticketsystemcontroller

FROM alpine:latest
COPY --from=build /ticketsystemcontroller/ticketsystemcontroller.bin /ticketsystemcontroller/ticketsystemcontroller.bin
EXPOSE 14801
ENTRYPOINT ["/ticketsystemcontroller/ticketsystemcontroller.bin"]
