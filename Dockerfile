# Originally modified from the main Cloud Run documentation

FROM golang:1.23.0-bookworm@sha256:31dc846dd1bcca84d2fa231bcd16c09ff271bcc1a5ae2c48ff10f13b039688f3 AS builder

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
