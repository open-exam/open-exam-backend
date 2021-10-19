FROM golang:1.17.2-buster AS builder
COPY . /open-exam-backend
WORKDIR /open-exam-backend
RUN go mod download
RUN go mod tidy
WORKDIR /open-exam-backend/user-service
RUN CGO_ENABLED=0 go build -o user-db-service -a -ldflags '-extldflags "-static"' src/*.go

# Bundle Stage
FROM alpine
COPY --from=builder /open-exam-backend/user-service/user-service .
COPY --from=builder /open-exam-backend/user-service/.prod.env .
EXPOSE 8080
ENTRYPOINT ["./user-service"]