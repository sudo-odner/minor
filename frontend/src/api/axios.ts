import axios from 'axios';

// Базовая конфигурация
export const api = axios.create({
  // baseURL: '/', // Мы за Traefik, поэтому корень (или /api/v1 если всё там)
  baseURL: '/',
  withCredentials: true, // ОБЯЗАТЕЛЬНО для передачи HttpOnly Cookies
  headers: {
    'Content-Type': 'application/json',
  },
});

// 1. ПЕРЕХВАТЧИК ЗАПРОСОВ (Добавляем токен в каждый запрос)
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('accessToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 2. ПЕРЕХВАТЧИК ОТВЕТОВ (Магия автоматического обновления токена)
api.interceptors.response.use(
  (response) => response, // Если ответ 2xx, просто пропускаем
  async (error) => {
    const originalRequest = error.config;

    // Если ошибка 401 (Unauthorized) и мы еще не пробовали обновлять токен (флаг _retry)
    if (error.response?.status === 401 && !originalRequest.url.includes('/refresh') && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        // Пытаемся получить новый Access Token
        // ВАЖНО: используем axios, а не нашу переменную api, чтобы не зациклиться
        const response = await axios.post(
          '/api/v1/auth/refresh',
          {},
          { withCredentials: true } // Обязательно прикладываем Refresh-куку
        );

        const { access_token } = response.data;

        // Сохраняем новый токен
        localStorage.setItem('accessToken', access_token);

        // Обновляем заголовок в упавшем запросе
        originalRequest.headers.Authorization = `Bearer ${access_token}`;

        // Повторяем изначальный запрос с новым токеном
        return api(originalRequest);
      } catch (refreshError) {
        // Если даже /refresh вернул ошибку (кука протухла или удалена)
        console.error('Refresh token expired or invalid');
        
        localStorage.removeItem('accessToken');
        
        // Только если мы не на странице логина, перенаправляем пользователя
        if (window.location.pathname !== '/login') {
          window.location.href = '/login';
        }
        
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);