FROM node:20-alpine

WORKDIR /app

# Копируем файлы зависимостей
COPY package*.json ./

# Устанавливаем зависимости
RUN npm install

# Копируем всё остальное
COPY . .

# Открываем порт Vite
EXPOSE 5173

# Запускаем с флагом --host, чтобы Vite слушал внешние подключения (от Traefik)
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]