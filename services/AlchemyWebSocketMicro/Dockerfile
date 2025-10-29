# Build stage
FROM golang:1.21-alpine AS builder

# Instalar certificados SSL y herramientas de build
RUN apk --no-cache add ca-certificates git

WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod ./
COPY go.sum* ./

# Descargar dependencias
RUN go mod download && go mod tidy

# Copiar código fuente
COPY . .

# Build de la aplicación con optimizaciones
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -a -installsuffix cgo \
    -ldflags='-w -s -extldflags "-static"' \
    -o main .

# Final stage - imagen mínima
FROM alpine:latest

# Instalar certificados SSL para conexiones HTTPS/WSS
RUN apk --no-cache add ca-certificates tzdata

# Crear usuario no-root para seguridad
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copiar el binario desde el builder
COPY --from=builder /app/main .

# Cambiar ownership al usuario no-root
RUN chown -R appuser:appgroup /app

# Cambiar a usuario no-root
USER appuser

# Exponer puerto
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8081/api/v1/health || exit 1

# Ejecutar la aplicación
CMD ["./main"]