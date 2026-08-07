# --- build ---
# Pin linux/amd64 so images built on Apple Silicon still run on Fargate (default X86_64).
FROM --platform=linux/amd64 golang:1.26-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/builder-api ./cmd/builder-api

# --- run ---
FROM --platform=linux/amd64 gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/builder-api /builder-api

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/builder-api"]
