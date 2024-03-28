# Build stage
FROM golang:latest AS build

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# Final stage
FROM alpine:latest

WORKDIR /root/
COPY --from=build /app/server .

# Create non-root user
RUN adduser -D -u 1001 emoteUser
# Change ownership of the binary to the non-root user
RUN chown emoteUser:emoteUser server
# Switch to the non-root user
USER emoteUser

EXPOSE 8089
CMD ["./server"]
