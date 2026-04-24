FROM golang:1.22-alpine AS build
WORKDIR /app

COPY go.mod ./
COPY main.go ./
RUN go build -o /out/headerdump .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/headerdump /app/headerdump

ENV PORT=8080
EXPOSE 8080

CMD ["/app/headerdump"]
