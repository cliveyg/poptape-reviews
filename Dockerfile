FROM golang:1.21-alpine AS build

RUN apk --no-cache add git

RUN mkdir /app
ADD . /app
WORKDIR /app

# remove any go module files and get deps
RUN rm -f go.mod go.sum
RUN go mod init github.com/cliveyg/poptape-reviews
RUN go mod tidy

RUN go mod download

# need these flags or alpine image won't run due to dynamically linked libs in binary
RUN CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build -a -ldflags '-w' -o reviews

FROM alpine:latest

RUN mkdir -p /reviews
COPY --from=build /app/reviews /reviews
COPY --from=build /app/.env /reviews
WORKDIR /reviews

# Make port 8020 available to the world outside this container
EXPOSE $PORT

# Run reviews binary when the container launches
CMD ["./reviews"]
