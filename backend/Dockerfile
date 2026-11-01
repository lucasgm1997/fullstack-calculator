# syntax=docker/dockerfile:1
# The build stage runs on the builder's native platform and cross-compiles
# (Go needs no emulation), so multi-arch images (amd64 + arm64) build fast.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
ENV GIN_MODE=release PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
