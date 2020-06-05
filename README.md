# goGenPdf
Сервис для генерации PDF, на вход принимает POST запрос в формате multipart/form-data, на выходе готовый pdf файл.

## Входные параметры:
1. **file** - файл html
1. **page_size** - размер страницы (пример А4)
1. **orientation** - ориентация страницы (пример Landscape)
1. **margin_top** - отступ сверху в пикселях (пример 10)
1. **margin_bottom** - отступ снизу в пикселях (пример 10)
1. **margin_left** - отступ слева в пикселях (пример 10)
1. **margin_right** - отступ справа в пикселях (пример 10)
1. **zoom** - увеличение/уменьшение (необязательный параметр, по умолчаню = 1, если больше 1 - увеличение, если меньше - уменьшение)
1. **shrink** - уменьшение размера, подгонка под размер страницы (необязательный параметр, по умолчаню = 0, 1 - включено, 0 - выключено)

## Пример запроса Curl:
```bash
curl -i \
-F "file=@test.html" \
-F "page_size=A4" \
-F "orientation=Landscape" \
-F "margin_bottom=7" \
-F "margin_top=7" \
-F "margin_left=10" \
-F "margin_right=10" \
-F "zoom=0.95" \
-F "shrink=1" \
http://localhost:8080
```

## Использование Docker
Создание контейнера
```bash
docker build -t go-genpdf .
или
docker-compose build
```
Запуск контейнера
```bash
docker run --publish 8080:8080 --name test --rm go-genpdf:latest
или
docker-compose up -d
```