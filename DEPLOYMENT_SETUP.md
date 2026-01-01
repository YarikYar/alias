# Настройка деплоя на новый сервер

## 1. Подготовка сервера

### Требования
- Ubuntu 20.04+ / Debian 11+
- Docker 24.0+
- Docker Compose v2.0+
- 2GB RAM минимум
- 20GB диск

### Установка Docker

```bash
# Обновить систему
sudo apt update && sudo apt upgrade -y

# Установить зависимости
sudo apt install -y ca-certificates curl gnupg lsb-release

# Добавить Docker GPG ключ
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# Добавить репозиторий
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Установить Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Добавить пользователя в группу docker
sudo usermod -aG docker $USER

# Перелогиниться
exit
```

### Проверка установки

```bash
docker --version
docker compose version
```

## 2. Настройка SSH для GitHub Actions

### На локальной машине

```bash
# Генерация SSH ключа
ssh-keygen -t rsa -b 4096 -C "github-actions-deploy" -f ~/.ssh/alias_deploy -N ""

# Показать приватный ключ (скопируй для GitHub Secrets)
cat ~/.ssh/alias_deploy

# Показать публичный ключ
cat ~/.ssh/alias_deploy.pub
```

### На сервере

```bash
# Создать директорию для SSH ключей
mkdir -p ~/.ssh
chmod 700 ~/.ssh

# Добавить публичный ключ
nano ~/.ssh/authorized_keys
# Вставь содержимое ~/.ssh/alias_deploy.pub
# Сохрани (Ctrl+O, Enter, Ctrl+X)

# Установить права
chmod 600 ~/.ssh/authorized_keys
```

### Проверка SSH доступа

```bash
# На локальной машине
ssh -i ~/.ssh/alias_deploy user@your-server-ip

# Если работает - отлично!
```

## 3. Подготовка директории на сервере

```bash
# Создать директорию для проекта
mkdir -p ~/alias
cd ~/alias

# Создать .env файл
nano .env
```

### Содержимое .env файла

```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=elias
DB_PASSWORD=СГЕНЕРИРУЙ_СЛОЖНЫЙ_ПАРОЛЬ_ЗДЕСЬ
DB_NAME=elias

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=СГЕНЕРИРУЙ_СЛОЖНЫЙ_ПАРОЛЬ_ЗДЕСЬ

# Telegram Bot
BOT_TOKEN=ТВОЙ_BOT_TOKEN_ОТ_BOTFATHER

# API URLs
API_URL=https://api.yourdomain.com
VITE_API_URL=https://api.yourdomain.com

# Traefik (если используется)
TRAEFIK_NETWORK=traefik_default

# Frontend domain
FRONTEND_DOMAIN=yourdomain.com

# Backend domain
BACKEND_DOMAIN=api.yourdomain.com
```

### Генерация паролей

```bash
# Сгенерировать случайные пароли
openssl rand -base64 32  # для DB_PASSWORD
openssl rand -base64 32  # для REDIS_PASSWORD
```

## 4. Настройка Traefik (опционально, но рекомендуется)

### Создать docker-compose для Traefik

```bash
mkdir -p ~/traefik
cd ~/traefik
nano docker-compose.yml
```

```yaml
version: '3.8'

services:
  traefik:
    image: traefik:v2.10
    container_name: traefik
    restart: unless-stopped
    security_opt:
      - no-new-privileges:true
    networks:
      - traefik_default
    ports:
      - 80:80
      - 443:443
    environment:
      - CF_API_EMAIL=${CF_API_EMAIL}
      - CF_DNS_API_TOKEN=${CF_DNS_API_TOKEN}
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./traefik.yml:/traefik.yml:ro
      - ./acme.json:/acme.json
      - ./config.yml:/config.yml:ro
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.traefik.entrypoints=http"
      - "traefik.http.routers.traefik.rule=Host(`traefik.yourdomain.com`)"
      - "traefik.http.middlewares.traefik-auth.basicauth.users=admin:$$apr1$$..."
      - "traefik.http.middlewares.traefik-https-redirect.redirectscheme.scheme=https"
      - "traefik.http.middlewares.sslheader.headers.customrequestheaders.X-Forwarded-Proto=https"
      - "traefik.http.routers.traefik.middlewares=traefik-https-redirect"
      - "traefik.http.routers.traefik-secure.entrypoints=https"
      - "traefik.http.routers.traefik-secure.rule=Host(`traefik.yourdomain.com`)"
      - "traefik.http.routers.traefik-secure.middlewares=traefik-auth"
      - "traefik.http.routers.traefik-secure.tls=true"
      - "traefik.http.routers.traefik-secure.tls.certresolver=cloudflare"
      - "traefik.http.routers.traefik-secure.tls.domains[0].main=yourdomain.com"
      - "traefik.http.routers.traefik-secure.tls.domains[0].sans=*.yourdomain.com"
      - "traefik.http.routers.traefik-secure.service=api@internal"

networks:
  traefik_default:
    external: true
```

### Создать traefik.yml

```bash
nano traefik.yml
```

```yaml
api:
  dashboard: true
  debug: true

entryPoints:
  http:
    address: ":80"
    http:
      redirections:
        entryPoint:
          to: https
          scheme: https
  https:
    address: ":443"

serversTransport:
  insecureSkipVerify: true

providers:
  docker:
    endpoint: "unix:///var/run/docker.sock"
    exposedByDefault: false
  file:
    filename: /config.yml

certificatesResolvers:
  cloudflare:
    acme:
      email: your-email@example.com
      storage: acme.json
      dnsChallenge:
        provider: cloudflare
        resolvers:
          - "1.1.1.1:53"
          - "1.0.0.1:53"
```

### Создать сеть и запустить

```bash
# Создать acme.json
touch acme.json
chmod 600 acme.json

# Создать config.yml
touch config.yml

# Создать сеть
docker network create traefik_default

# Запустить Traefik
docker compose up -d
```

## 5. Настройка GitHub Secrets

Перейди в настройки репозитория на GitHub:
`https://github.com/YarikYar/alias/settings/secrets/actions`

### Добавь следующие секреты:

| Имя секрета | Значение | Откуда взять |
|-------------|----------|--------------|
| `SSH_PRIVATE_KEY` | Приватный SSH ключ | `cat ~/.ssh/alias_deploy` |
| `SERVER_HOST` | IP адрес сервера | `123.45.67.89` |
| `SERVER_USER` | SSH пользователь | `ubuntu` или твой user |
| `DEPLOY_PATH` | Путь к проекту | `/home/ubuntu/alias` |

### Как добавить секрет:

1. Нажми "New repository secret"
2. Введи Name (например, `SSH_PRIVATE_KEY`)
3. Вставь Value (содержимое ключа/значение)
4. Нажми "Add secret"

## 6. Первый деплой

### Вариант 1: Автоматический (через GitHub Actions)

```bash
# На локальной машине
git push origin main

# GitHub Actions автоматически задеплоит на сервер
# Проверь статус: https://github.com/YarikYar/alias/actions
```

### Вариант 2: Ручной

```bash
# На сервере
cd ~/alias
git clone git@github.com:YarikYar/alias.git .
# или если уже есть: git pull origin main

# Сделать deploy.sh исполняемым
chmod +x deploy.sh

# Запустить деплой
./deploy.sh
```

## 7. Проверка работы

### Проверить контейнеры

```bash
cd ~/alias
docker compose -f docker-compose.prod.yml ps
```

Должно быть:
- ✅ alias-backend-1 (Up)
- ✅ alias-frontend-1 (Up)
- ✅ alias-postgres-1 (Up, healthy)
- ✅ alias-redis-1 (Up, healthy)

### Проверить логи

```bash
# Все логи
docker compose -f docker-compose.prod.yml logs -f

# Только backend
docker compose -f docker-compose.prod.yml logs -f backend

# Только frontend
docker compose -f docker-compose.prod.yml logs -f frontend
```

### Проверить API

```bash
# Health check
curl http://localhost:8080/health

# Или через домен (если настроен Traefik)
curl https://api.yourdomain.com/health
```

### Проверить frontend

```bash
# Через браузер
# http://your-server-ip (если без Traefik)
# https://yourdomain.com (если с Traefik)
```

## 8. Настройка DNS

Если используешь Traefik с SSL:

### Добавь A записи:

| Тип | Имя | Значение | TTL |
|-----|-----|----------|-----|
| A | @ | IP_СЕРВЕРА | 300 |
| A | api | IP_СЕРВЕРА | 300 |
| A | * | IP_СЕРВЕРА | 300 |

### Проверка DNS

```bash
# Проверить резолв
nslookup yourdomain.com
nslookup api.yourdomain.com

# Должен показывать IP твоего сервера
```

## 9. Настройка firewall

### UFW (Ubuntu Firewall)

```bash
# Установить UFW
sudo apt install -y ufw

# Разрешить SSH
sudo ufw allow 22/tcp

# Разрешить HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Включить firewall
sudo ufw enable

# Проверить статус
sudo ufw status
```

## 10. Мониторинг и обслуживание

### Автоматический перезапуск при сбое

Docker Compose уже настроен с `restart: unless-stopped`

### Логи с ротацией

```bash
# Настроить Docker logging
sudo nano /etc/docker/daemon.json
```

```json
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```

```bash
# Перезапустить Docker
sudo systemctl restart docker
```

### Бэкапы базы данных

```bash
# Создать директорию для бэкапов
mkdir -p ~/backups

# Создать скрипт бэкапа
nano ~/backup.sh
```

```bash
#!/bin/bash
BACKUP_DIR=~/backups
DATE=$(date +%Y%m%d_%H%M%S)
CONTAINER=alias-postgres-1

docker exec $CONTAINER pg_dump -U elias elias | gzip > $BACKUP_DIR/backup_$DATE.sql.gz

# Удалить бэкапы старше 7 дней
find $BACKUP_DIR -name "backup_*.sql.gz" -mtime +7 -delete

echo "Backup completed: backup_$DATE.sql.gz"
```

```bash
# Сделать исполняемым
chmod +x ~/backup.sh

# Добавить в cron (каждый день в 3:00)
crontab -e
# Добавить строку:
0 3 * * * /home/ubuntu/backup.sh >> /home/ubuntu/backup.log 2>&1
```

## 11. Troubleshooting

### Контейнеры не запускаются

```bash
# Проверить логи
docker compose -f docker-compose.prod.yml logs

# Проверить .env файл
cat .env

# Пересоздать контейнеры
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d --force-recreate
```

### GitHub Actions деплой не работает

```bash
# Проверить секреты в GitHub
# Проверить SSH доступ вручную
ssh -i ~/.ssh/alias_deploy user@server

# Проверить логи Actions в GitHub
# https://github.com/YarikYar/alias/actions
```

### База данных не подключается

```bash
# Проверить контейнер
docker ps -a | grep postgres

# Проверить логи
docker logs alias-postgres-1

# Подключиться вручную
docker exec -it alias-postgres-1 psql -U elias -d elias
```

### SSL сертификат не выдается

```bash
# Проверить Traefik логи
docker logs traefik

# Проверить DNS
nslookup yourdomain.com

# Проверить acme.json
cat ~/traefik/acme.json

# Пересоздать acme.json
rm ~/traefik/acme.json
touch ~/traefik/acme.json
chmod 600 ~/traefik/acme.json
docker restart traefik
```

## 12. Обновление приложения

### Автоматическое (через GitHub)

```bash
# На локальной машине
git add .
git commit -m "feat: new feature"
git push origin main

# GitHub Actions автоматически задеплоит
```

### Ручное

```bash
# На сервере
cd ~/alias
git pull origin main
./deploy.sh
```

## 13. Откат к предыдущей версии

```bash
# На сервере
cd ~/alias

# Посмотреть коммиты
git log --oneline

# Откатиться к нужному коммиту
git checkout <commit-hash>

# Задеплоить
./deploy.sh

# Вернуться на latest
git checkout main
./deploy.sh
```

## 14. Полезные команды

```bash
# Статус всех контейнеров
docker ps -a

# Использование ресурсов
docker stats

# Очистка неиспользуемых образов
docker system prune -a

# Рестарт всех сервисов
docker compose -f docker-compose.prod.yml restart

# Остановка всех сервисов
docker compose -f docker-compose.prod.yml down

# Запуск всех сервисов
docker compose -f docker-compose.prod.yml up -d

# Просмотр логов в реальном времени
docker compose -f docker-compose.prod.yml logs -f --tail=100
```

## 15. Чеклист после деплоя

- [ ] Все контейнеры запущены и healthy
- [ ] API отвечает на health check
- [ ] Frontend загружается
- [ ] WebSocket подключается
- [ ] Игра работает end-to-end
- [ ] SSL сертификат выдан (если используется)
- [ ] DNS настроен корректно
- [ ] Firewall настроен
- [ ] Бэкапы настроены
- [ ] Мониторинг работает
- [ ] Логи пишутся корректно

## Готово! 🎉

Твое приложение теперь задеплоено на сервере с автоматическим CI/CD через GitHub Actions.

При каждом пуше в `main` ветку будет происходить автоматический деплой на сервер.

---

**Поддержка:**
- GitHub Issues: https://github.com/YarikYar/alias/issues
- Документация: см. README.md
