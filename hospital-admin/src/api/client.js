import axios from 'axios';

const api = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
});

// JWT interceptor — tự động gắn token vào mọi request
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Response interceptor — xử lý lỗi chung
api.interceptors.response.use(
  (response) => {
    const code = response.data?.code;
    if ([3001, 3002, 3009].includes(code)) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      localStorage.setItem('auth_error', response.data?.message || 'Phiên đăng nhập không hợp lệ');
      window.location.href = '/login';
    }
    return response;
  },
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export default api;
