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

# ---> ADICIONE ESTA LINHA PARA DAR PERMISSÃO <---
RUN chmod +x ./main && chmod -R 777 /root

EXPOSE 8080

ENTRYPOINT ["./main"]