# Telegram-бот на Go для управления сериалами через Sonarr
Этот Telegram-бот создан для управления сериалами через Sonarr. С его помощью можно искать, добавлять, редактировать, удалять и упорядочивать сериалы в вашей библиотеке Sonarr. Бот написан на Go, потребляет менее 10 МБ оперативной памяти, не хранит состояние на диске (кроме журналов ошибок), а размер Docker-образа не превышает 10 МБ в сжатом виде. Workflow собирает образ только для `x86_64`/`amd64`. Для других архитектур (`arm64`, `arm32v7`) верните их в параметр `platforms` в `.github/workflows/main.yaml` и добавьте шаг `docker/setup-qemu-action@v3`.

Бот построен на [golift/starr](https://github.com/golift/starr/) и [go-telegram-bot-api/telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api/) без каких-либо дополнительных зависимостей. Интерфейс бота полностью переведён на русский язык.

## Возможности и команды

<img src="screenshots/menu.png?raw=true" alt="menu" title="menu" width="300" />

### Запуск бота
<img src="screenshots/start.png?raw=true" alt="start" title="start" width="300" />

### Поиск и добавление сериалов
``/q [сериал]`` или просто название сериала в сообщении: поиск сериала.\
Когда сериал найден, бот предлагает добавить его в библиотеку Sonarr с разными настройками отслеживания. Если у вас одна корневая папка и один профиль качества, бот выберет их автоматически. Если вариантов несколько, бот попросит выбрать корневую папку и профиль качества. Если в Sonarr заданы теги, их тоже можно выбрать.

<img src="screenshots/add_links.png?raw=true" alt="qlinks" title="add series" width="300" />
<img src="screenshots/add_confirmation.png?raw=true" alt="qconfirmation" title="add confirmation" width="300" />
<img src="screenshots/add_monitor.png?raw=true" alt="qmonitoring" title="add monitoring" width="300" />
<img src="screenshots/add_search.png?raw=true" alt="qsearch" title="add search" width="300" />

### Управление сериалами
``/library [сериал]`` или ``/l [сериал]``: управление сериалами в библиотеке. Можно изменить профиль качества сериала (если в Sonarr настроено больше одного) и теги. Также можно включать и отключать отслеживание, удалять сериал, запускать его поиск, изменять, удалять и искать сезоны, а также смотреть занятое место на диске. Название необязательно: если его не указать, откроется меню фильтров.

<img src="screenshots/library.png?raw=true" alt="l" title="library" width="300" />
<img src="screenshots/library_series.png?raw=true" alt="lseries" title="library series" width="300" />
<img src="screenshots/library_seasons.png?raw=true" alt="lseasons" title="library seasons" width="300" />
<img src="screenshots/library_season.png?raw=true" alt="lseason" title="library season" width="300" />

### Удаление сериалов
``/delete [сериал]`` или ``/d [сериал]``: удаление одного или нескольких сериалов из библиотеки Sonarr. Будьте осторожны: вместе с сериалом удаляются и связанные файлы. Название необязательно: если его не указать, все сериалы будут показаны кнопками, и можно выбрать несколько.

<img src="screenshots/delete_confirmation.png?raw=true" alt="q1" title="delete" width="300" />

### Отмена команд
``/clear`` или ``/cancel`` или ``/stop``:
сбрасывает все ранее введённые команды и состояние бота. Можно использовать в любой момент.

### Управление библиотекой
- ``/up`` или ``/upcoming``: список серий, которые выйдут в ближайшие 30 дней
- ``/rss``: запустить RSS-синхронизацию
<!---
- ``/searchmonitored``: поиск всех отслеживаемых сериалов и серий
- ``/updateall``: обновить метаданные и пересканировать файлы и папки для всех сериалов и серий
-->

### Информация о системе
- ``/free`` или ``/diskspace``: свободное место на дисках, подключённых к серверу Sonarr
- ``/system``: конфигурация Sonarr
- ``/id`` или ``/getid``: показать ваш Telegram ID

## Установка и настройка
Бота можно собрать самостоятельно из исходного кода или использовать Docker-образ из GitHub Container Registry и Docker Hub:
- GitHub [ghcr.io/woiza/telegram-bot-sonarr](https://github.com/woiza/telegram-bot-sonarr/pkgs/container/telegram-bot-sonarr)
- Docker Hub [woiza/telegram-bot-sonarr](https://hub.docker.com/repository/docker/woiza/telegram-bot-sonarr/)

Для работы бота нужно задать семь обязательных переменных окружения. Подробности смотрите в примере Docker Compose ниже. Перед запуском получите токен Telegram-бота и API-ключ Sonarr. Также определите, кому разрешён доступ к боту (ID пользователя Telegram). Можно указать несколько пользователей списком ID. Как получить эти данные, описано в официальной документации:
- [Токен Telegram-бота](https://core.telegram.org/bots/tutorial/)
- [API-ключ Sonarr](https://wiki.servarr.com/en/sonarr/settings#security/)



### Сборка Docker-образа
```
docker buildx build --push --platform linux/amd64,linux/arm64,linux/arm/v7 --tag <repo>/<image>:<tag> .
```


### Пример Docker Compose
```
services:
    telegram-bot-sonarr:
        image: woiza/telegram-bot-sonarr
        mem_limit: 128M
        container_name: telegram-bot-sonarr
#        depends_on:
#            - sonarr
        restart: always
        environment:
            - SBOT_TELEGRAM_BOT_TOKEN=1460...:AAHlBW_mabVg...
            - SBOT_BOT_ALLOWED_USERIDS=123,987,-567 # ID пользователей Telegram, ID групп отрицательные
            - SBOT_BOT_MAX_ITEMS=10 # количество элементов на странице
            - SBOT_BOT_IGNORE_TAGS=false # true/false; true = бот не будет спрашивать теги (удобно при автотегировании)
            - SBOT_BOT_SERIES_TYPE= # необязательно, возможные значения: standard, daily, anime. Если задано, бот не будет спрашивать тип сериала
            - SBOT_SONARR_PROTOCOL=http # http или https
            - SBOT_SONARR_PORT=8989
            - SBOT_SONARR_HOSTNAME=192.168.2.2 # IP-адрес или имя хоста
            - SBOT_SONARR_BASE_URL= # необязательно, например /sonarr, в зависимости от настроек Sonarr
            - SBOT_SONARR_API_KEY=1010d7...
            - SBOT_WEBHOOK_PORT=8090 # необязательно, приём уведомлений Sonarr
            - SBOT_WEBHOOK_TOKEN=секрет # необязательно, секрет в адресе вебхука
```
### Команды для /setcommands в BotFather

```
q - поиск сериала
library - список всех сериалов (ВНИМАНИЕ: может быть большим)
delete - удаление сериала (ВНИМАНИЕ: список может быть большим)
clear - сбросить все ранее отправленные команды
free - свободное место на дисках
up - серии, которые выйдут в ближайшие 30 дней
rss - запустить RSS-синхронизацию
system - конфигурация Sonarr
id - показать ваш Telegram ID
```

## Уведомления на русском языке
Sonarr сам отправляет уведомления в Telegram только на английском, и перевести их настройками нельзя. Поэтому бот умеет принимать события Sonarr через подключение **Webhook** и отправлять их в Telegram на русском: начало загрузки, импорт, замена файла, добавление и удаление, проблемы со здоровьем системы, обновление приложения и другие.

1. Добавьте в окружение бота переменные (если не задан порт, приём событий отключён):
   - `SBOT_WEBHOOK_PORT=8090` - порт, на котором бот принимает события
   - `SBOT_WEBHOOK_TOKEN=секрет` - необязательный секрет, который проверяется в адресе
2. В Sonarr: *Settings → Connect → + → Webhook*. URL: `http://telegram-bot-sonarr:8090/webhook?token=секрет`, метод **POST**. Отметьте нужные события и нажмите *Test*.

Уведомления отправляются всем пользователям из `SBOT_BOT_ALLOWED_USERIDS`. Тексты проблем со здоровьем системы (Sonarr формирует их сам) приходят как есть, на английском.

## Участие в разработке
Вы можете помочь проекту: отправляйте pull request'ы, сообщайте об ошибках и предлагайте улучшения. Любой вклад приветствуется!

## Пиво
Если вам нравится наша работа, угостите нас освежающим напитком.

<a href="https://paypal.me/telegramarrbots?country.x=EUR" target="_blank">
  <img src="pp.png?raw=true" alt="q1" title="donate" width="200">
</a>

## Лицензия
Этот Telegram-бот распространяется по [лицензии MIT](https://opensource.org/license/mit/).
