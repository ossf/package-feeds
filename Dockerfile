# Originally modified from the main Cloud Run documentation

# golang:1.20.2-buster
FROM golang@sha256:23050c2510e0a920d66b48afdc40043bcfe2e25d044a2d7b33475632d83ab6c7 as builder

# Create and change to the app directory.
WORKDIR /app
ENV CGO_ENABLED=0
# # Retrieve application dependencies.
# # This allows the container build to reuse cached dependencies.
# # Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
RUN go build -mod=readonly -v -o server ./cmd/scheduled-feed

#https://github.com/GoogleContainerTools/distroless/
FROM gcr.io/distroless/base:nonroot

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /app/server

# Run the web service on container startup.
CMD ["/app/server"]
