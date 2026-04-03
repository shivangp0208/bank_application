# building stage
FROM golang:1.26.1-alpine3.23 AS build
WORKDIR /app
COPY . .
RUN go build -o main main.go

# running stage 
FROM alpine:3.23
WORKDIR /app
COPY --from=build /app/main .
COPY config.yaml .

EXPOSE 8080
CMD [ "/app/main" ]