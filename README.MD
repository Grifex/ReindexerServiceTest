# Как запускать yaml конфигурацией

```bash
docker compose up -d redis reindexer

go run ./cmd/api/main.go

```

# Как запускать с env конфигурацией


```bash

docker compose up -d --build

```

# API 

-POST /documents — создать документ

-GET /documents?limit=20&offset=0 — список документов (пагинация) (где limit кол-во документов на страницу, offset - номер страницы)

-GET /documents/{id} — получить один документ

-PUT /documents/{id} — обновить документ

-DELETE /documents/{id} — удалить документ

# Пример данных для запросов

```
    {
    "title": "Название документа 1",
    "remove-field": "удаляемое поле 1",
    "level1": [
      {
        "sort": 1,
        "name": "номер 1",
        "remove-field": "удаляемое поле 1",
        "level2": [
          {"key": "k1", "value": "v1", "remove-field": "удаляемое поле 1"},
          {"key": "k2", "value": "v2", "remove-field": "удаляемое поле 1"}
        ]
      },
      {
        "sort": 10,
        "name": "номер 2",
        "remove-field": "удаляемое поле 1",
        "level2": [
          {"key": "k3", "value": "v3", "remove-field": "удаляемое поле 1"}
        ]
      },
      {
        "sort": 20,
        "name": "номер 3",
        "remove-field": "удаляемое поле 1",
        "level2": [
          {"key": "k4", "value": "v4", "remove-field": "удаляемое поле 1"}
        ]
      }
    ]
  }

```