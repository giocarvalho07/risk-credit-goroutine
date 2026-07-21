FROM alpine:3.19

WORKDIR /app

# Copia o binário e os arquivos de configuração necessários
COPY main .
COPY config.yaml .
COPY internal/data ./internal/data

# Garante permissões de execução e leitura totais
RUN chmod +x ./main && chmod -R 777 /app

EXPOSE 8080

CMD ["./main"]