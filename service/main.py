import os
from fastapi import FastAPI
from sqlalchemy import text
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker

# Получение строки подключения к БД из переменных окружения 
DATABASE_URL = os.getenv("DATABASE_URL")

# Создание асинхронного движка для работы с БД. 
# echo=False отключает автоматическое логирование всех SQL-запросов в консоль
engine = create_async_engine(DATABASE_URL, echo=False)

# Создание фабрики сессий. expire_on_commit=False предотвращает принудительное 
# обновление объектов после коммита (важно для асинхронного режима)
Session = async_sessionmaker(engine, expire_on_commit=False)

# Инициализация приложения FastAPI с указанием его названия
app = FastAPI(title="service")

# Обработчик события старта приложения
@app.on_event("startup")
async def init_db():
    """Создание таблицы при старте, если её нет."""
    # engine.begin() автоматически открывает транзакцию и делает COMMIT при выходе из контекста
    async with engine.begin() as conn:
        # Выполнение сырого SQL-запроса для создания таблицы для демонстрации/тестов
        await conn.execute(text("""
            CREATE TABLE IF NOT EXISTS items (
                id SERIAL PRIMARY KEY,
                name TEXT NOT NULL,
                created_at TIMESTAMPTZ DEFAULT NOW()
            )
        """))


# Эндпоинт проверки работоспособности сервиса (Health check)
@app.get("/health")
async def health():
    return {"status": "ok"}


# Эндпоинт для получения списка элементов
@app.get("/items")
async def list_items():
    """Вернуть все записи из БД."""
    # Открытие асинхронной сессии для выполнения запросов
    async with Session() as session:
        # Асинхронное выполнение SQL-запроса 
        result = await session.execute(text("SELECT id, name FROM items ORDER BY id"))
        # Извлечение всех строк из результата
        rows = result.fetchall()
        # Преобразование кортежей БД в список словарей для автоматической конвертации в JSON
        return [{"id": r[0], "name": r[1]} for r in rows]


# Эндпоинт для добавления нового элемента через URL-параметр (Path parameter)
@app.post("/items/{name}")
async def add_item(name: str):
    """Добавить запись в БД."""
    async with Session() as session:
        # Использование именованных плейсхолдеров для защиты от SQL-инъекций
        result = await session.execute(
            text("INSERT INTO items (name) VALUES (:name) RETURNING id"),
            {"name": name},
        )
        # Получение значения первого поля первой строки (сгенерированный id)
        new_id = result.scalar()
        # Сессия требует явного подтверждения транзакции
        await session.commit()
        return {"id": new_id, "name": name}