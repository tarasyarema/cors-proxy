FROM golang:latest as builder

WORKDIR /src

COPY go.* ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o server

FROM alpine:latest
RUN apk add --no-cache ca-certificates

COPY --from=builder /src/server /server

# Run the web service on container startup.
CMD ["/server"]