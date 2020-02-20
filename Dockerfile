FROM golang:latest as builder
LABEL maintainer="schneider@puzzle.ch,buehlmann@puzzle.ch"
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o operator .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
USER 11
WORKDIR /home/operator
COPY --from=builder /app/operator .
CMD ["./operator"]
