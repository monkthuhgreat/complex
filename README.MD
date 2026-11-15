# СТЕК: Golang + Python + PostgreSQL

## АРХИТЕКТУРА
- **gateway** (порт 8080) - точка входа, проксирует запросы в **service**
- **service** (порт 8001) - логика, работает с БД
- **postgres** (порт 5432) - хранилище

## ЗАПУСК

### Вариант 1: Локально через Docker Compose 
```bash
docker compose up --build
```

Приложение будет доступно по адресу `http://localhost:8080`.

---

### Вариант 2: Развертывание в кластере Kubernetes (k3s)

1. Скопируйте `.env.example` в `.env` и заполните переменные окружения 
2. Запустите скрипт автоматической сборки и деплоя:

```bash
chmod +x scripts/start.sh
./scripts/start.sh
```