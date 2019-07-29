FROM golang as builder
COPY . /go/src/github.com/flaccid/
WORKDIR /go/src/github.com/flaccid/vsync
RUN go get ./... && \
    CGO_ENABLED=0 GOOS=linux go build -o /vsync cmd/vsync/vsync.go

FROM gcr.io/distroless/static
COPY --from=builder /vsync /vsync
ENTRYPOINT ["/vsync"]
