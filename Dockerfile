# Build Stage: Compile the Go Plugin
FROM golang:1.24.1-alpine AS builder

# Install necessary dependencies
RUN apk add --no-cache make cmake gcc g++ libc-dev

# Set working directory inside container
WORKDIR /app

# Copy Go source files
COPY . .

# Build the Go plugin as a shared object (.so)
RUN go build -buildmode=c-shared -o out_cloudant.so *.go

# Final Stage: Create Fluent Bit Image with Plugin
FROM fluent/fluent-bit:latest

# Set working directory
WORKDIR /fluent-bit

# Copy the compiled plugin from the builder stage
COPY --from=builder /app/out_cloudant.so /fluent-bit/bin/out_cloudant.so

# Set environment variable for Fluent Bit to use the plugin
ENV FLB_PLUGIN_PATH="/fluent-bit/bin/out_cloudant.so"

# Run Fluent Bit with the custom output plugin
ENTRYPOINT ["/fluent-bit/bin/fluent-bit", "-e", "/fluent-bit/bin/out_cloudant.so", "-c", "/fluent-bit/etc/fluent-bit.conf"]
