# build stage
FROM golang as builder
# Add dependencies
WORKDIR /go/src/app
ADD . /go/src/app
# Build app
RUN go mod download
RUN go build -o /go/bin/app ./cmd

# final stage
FROM gcr.io/distroless/base

COPY --from=builder /go/bin/app /
ENTRYPOINT ["/app"]