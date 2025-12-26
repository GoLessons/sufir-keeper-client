# Sufir's Keeper Client

## Пользователям

@todo

## Разработчикам

**Dev CA и Docker**
- По умолчанию путь к dev CA: `./var/ca.crt`. Этот путь используется CLI, если не задано явным образом.
- В Docker контейнере путь сохраняется тем же (`/workspace/var/ca.crt`) при маунте директории `./var` в `/workspace/var`.
- Пример маунта уже задан в `docker-compose.yml`. Создайте файл `./var/ca.crt` на хосте и он будет доступен в контейнере.
- Запуск контейнера: `docker compose up -d`
- Выполнение команд CLI внутри контейнера:
  - `docker compose exec go keepcli --ca-cert-path /workspace/var/ca.crt --server https://localhost:8443/api/v1 login`
  - Можно опустить флаг `--ca-cert-path`, так как установлен дефолт: `./var/ca.crt` → `/workspace/var/ca.crt`.
- Альтернатива через ENV: `docker compose exec go env SUFIR_KEEPER_CA_CERT=/workspace/var/ca.crt keepcli login`

**Примечания**
- Ожидается, что сервер поддерживает HTTP/2, клиент автоматически пытается использовать его.
- Все тесты и проверки запускаются внутри Docker.
