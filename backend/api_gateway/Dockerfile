FROM traefik:v3.1

# Копируем статический файл конфигурации
COPY traefik.yml /etc/traefik/traefik.yml

# Если динамическая конфигурация из файла
# COPY dynamic_conf.yml /etc/traefik/dynamic_conf.yml

# Порты для HTTP, HTTPS и Dashboard
EXPOSE 80 443 8080