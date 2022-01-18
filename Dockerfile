FROM golang:1.17-alpine
WORKDIR /app
RUN go install github.com/cespare/reflex@latest
COPY . .
RUN go build -o ./feature-flagger-api .
CMD ["./feature-flagger-api"]