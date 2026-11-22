# gRPC API

Proto: [proto/shortener/v1/shortener.proto](proto/shortener/v1/shortener.proto).

Сгенерировать Go-код:

```bash
make proto
```

gRPC слушает адрес из `GRPC_ADDRESS` / флага `-grpc` / поля `grpc_address` в JSON. Если не задан — порт HTTP + 1000 (например `:8080` → `:9080`).

Авторизация: metadata `authorization` с JWT (как cookie `Token`).