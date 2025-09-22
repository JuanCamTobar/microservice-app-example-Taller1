# =========================================
# Stage 1: Builder (compila el binario)
# =========================================
FROM golang:1.20-alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=off

RUN apk add --no-cache git ca-certificates wget && update-ca-certificates

# Instalar dep
RUN wget -qO /usr/local/bin/dep \
      https://github.com/golang/dep/releases/download/v0.5.4/dep-linux-amd64 \
    && chmod +x /usr/local/bin/dep

# Trabajar dentro del GOPATH (dep lo requiere)
WORKDIR /go/src/app

# 1) Cache base con lock actual
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only -v

# 2) Copiar código
COPY . .

# 3) **Actualizar vendor** en función de imports NUEVOS (gobreaker, etc.)
RUN dep ensure -v

# 4) Compilar
RUN mkdir -p /out
RUN go build -trimpath -ldflags="-s -w" -o /out/auth-api ./.

# =========================================
# Stage 2: Runtime (imagen mínima y segura)
# =========================================
FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app
COPY --from=builder /out/auth-api /usr/local/bin/auth-api

ENV PORT=8080
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/auth-api"]