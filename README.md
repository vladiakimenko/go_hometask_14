# Системные требования
go 1.22 или выше

# Структура проекта

```
.
├── cmd/
│   └── urlshortener/
│       └── main.go          # Точка входа, запуск сервера
├── internal/
│   ├── api/
│   │   ├── handlers.go       # Хендлеры
│   │   ├── handlers_test.go  # httptest тесты
│   │   ├── middleware.go     # Мидлвара
│   │   ├── models.go         # Модели для респонсов
│   │   └── utils.go          # Переиспользуемые хелперы
│   └── shortener/
│       ├── shortener.go      # Бизнес логика
│       └── shortener_test.go # Изолированные тесты бизнес логики
├── go.mod                    # Модуль
└── README.md                 # Этот файл
```


# Подробное описание задачи

### Функциональные требования:
Сервис должен предоставлять два HTTP-эндпоинта:  

POST /shorten - сокращение URL:  
· принимает JSON: {“url”: “example.com.../long/path”};  
· возвращает JSON: {“short_url”: “abc123”, “original_url”: “example.com.../long/path”};  
· генерирует уникальный короткий идентификатор (6-8 символов).  

GET /{short_url} - получение оригинального URL:  
· принимает короткий идентификатор в пути;  
· возвращает HTTP 302 редирект на оригинальный URL;  
· при отсутствии URL возвращает 404.  

### Технические требования:
· Хранение данных в памяти (map).  
· Валидация входящих URL (должны быть корректными HTTP/HTTPS адресами).  
· Обработка ошибок (некорректный JSON, невалидный URL, отсутствующий short_url).  
· Генерация уникальных коротких идентификаторов.  

### Обязательные конструкции:
· Table-driven tests для тестирования бизнес-логики.  
· httptest для тестирования HTTP-обработчиков.  
· t.Run для организации под-тестов.  
· Покрытие тестами не менее 80% для бизнес-логики.  


# Запуск приложения

```
go run cmd/urlshortener/main.go
```

# Примеры запросов

```
curl -i \
  -X POST http://localhost:8080/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/some/path"}'
```

Забираем 'short_url' из ответа и подставляем в 

```
curl -i \
  http://localhost:8080/{short_url}
```

# Запуск тестов

```
go test ./...
```

### Анализ покрытия
```
go test ./... -cover
```

```
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```


urlshortener/internal/api: 80.4%  
urlshortener/internal/shortener: 92.3%  