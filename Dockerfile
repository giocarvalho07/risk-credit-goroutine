# --- ETAPA DE BUILD (Exemplo) ---
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# --- ETAPA FINAL (Onde está o problema) ---
FROM alpine:latest  # ou gcr.io/distroless/static, debian, etc.
WORKDIR /app

# 1. Copia o binário
COPY --from=builder /app/main .

# 2. CORREÇÃO: Garante permissão de execução para qualquer usuário
RUN chmod +x /app/main

# (Opcional, mas altamente recomendado para OpenShift)
# Mudar a posse do arquivo e usar um usuário não-root explicitamente
RUN chmod -R 777 /app

EXPOSE 8080

ENTRYPOINT ["./main"]