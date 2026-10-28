# СТЕК: Golang + Python + PostgreSQL

## АРХИТЕКТУРА
- **gateway** (порт 8080) - точка входа, проксирует запросы в **service**
- **service** (порт 8001) - логика, работает с БД
- **postgres** (порт 5432) - хранилище

## ЗАПУСК
```bash
docker compose up --build