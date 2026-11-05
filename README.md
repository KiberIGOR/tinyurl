# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m v2 template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/v2 .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Бенчмарки

```bash
go test -bench=. -benchmem ./internal/repository/ ./internal/service/ ./internal/handler/
```

| Бенчмарк | Что измеряет |
|----------|----------------|
| `BenchmarkMemorySave/Get/...` | in-memory хранилище |
| `BenchmarkShorten/Resolve` | сервис сокращения |
| `BenchmarkPostURLHandler/GetURLHandler` | HTTP-хендлеры (httptest) |

## Профилирование памяти и нагрузка hey

1. Установите [hey](https://github.com/rakyll/hey): `go install github.com/rakyll/hey@latest`
2. Снимите **base**-профиль (до оптимизаций) или используйте готовый `profiles/base.pprof`
3. После оптимизаций — `profiles/result.pprof`
4. Нагрузка + heap с живого сервера (pprof на `:6060`):

```bash
./scripts/load_and_profile.sh profiles/result.pprof
# другой порт, если 8099 занят: SERVER_ADDR=:8101 ./scripts/load_and_profile.sh profiles/result.pprof
```

Скрипт поднимает сервер, прогревает cookie, гоняет **hey** на `POST /` и `GET /{id}`, сохраняет `http://localhost:6060/debug/pprof/heap?gc=1`.

Профили с бенчмарка `BenchmarkShorten`:

```bash
go test -bench=BenchmarkShorten -benchmem -benchtime=3s -memprofile=profiles/base.pprof ./internal/service/
go test -bench=BenchmarkShorten -benchmem -benchtime=3s -memprofile=profiles/result.pprof ./internal/service/
go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
```

### Оптимизации

- `generateID`: `sync.Pool` для буферов random/base64 вместо `EncodeToString` на каждый вызов
- `DeleteByUserID`: предвыделение слайса `cap(shortURLCh)`
- `audit.FileObserver`: один открытый файл вместо `OpenFile` на каждое событие

### Diff `pprof` (base → result)

```
File: service.test
Type: alloc_space
Showing nodes accounting for -356.30MB, 8.18% of 4355.75MB total
      flat  flat%   sum%        cum   cum%
 -117.77MB  2.70%  2.70%  -117.77MB  2.70%  github.com/KiberIGOR/tinyurl/internal/repository.(*Memory).Save
 -104.01MB  2.39%  5.09%   -97.51MB  2.24%  net/url.(*URL).joinPath
  -74.51MB  1.71%  6.80%   -74.51MB  1.71%  net/url.parse
     -46MB  1.06%  7.86%      -46MB  1.06%  encoding/base64.(*Encoding).EncodeToString
  -21.50MB  0.49%  8.35%      -20MB  0.46%  fmt.Sprintf
      12MB  0.28%  8.54%      -34MB  0.78%  github.com/KiberIGOR/tinyurl/internal/service.generateID
       9MB  0.21%  8.33%  -354.80MB  8.15%  github.com/KiberIGOR/tinyurl/internal/service.BenchmarkShorten
```

Отрицательные значения в колонке `flat` — меньше аллокаций в **result** относительно **base** (в т.ч. исчезновение `EncodeToString` в hot path `generateID`).

Полный вывод: `profiles/pprof_diff_top.txt`.
