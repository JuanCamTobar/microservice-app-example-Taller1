# =========================================
# Stage 1: Builder (compila el binario)
# =========================================
FROM golang:1.20-alpine AS builder

# Compilación clásica (GOPATH) para proyectos con dep/Gopkg.*
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=off

# Herramientas necesarias
RUN apk add --no-cache git ca-certificates wget && update-ca-certificates

# Instalar dep desde release binaria (evita usar go install con módulos)
RUN wget -qO /usr/local/bin/dep \
      https://github.com/golang/dep/releases/download/v0.5.4/dep-linux-amd64 \
    && chmod +x /usr/local/bin/dep

# Trabajamos dentro del GOPATH para modo GOPATH
WORKDIR /go/src/app

# --- Caché de dependencias ---
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -vendor-only -v

# --- Copiar el resto y compilar ---
COPY . .
# Compila el binario estático
RUN go build -trimpath -ldflags="-s -w" -o /out/auth-api ./.

# =========================================
# Stage 2: Runtime (imagen mínima y segura)
# =========================================
FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app

# Copiamos el binario
COPY --from=builder /out/auth-api /usr/local/bin/auth-api

# Puerto por defecto (ajústalo si tu main usa otro)
ENV PORT=8080
EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/auth-api"]