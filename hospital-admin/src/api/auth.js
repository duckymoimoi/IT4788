import api from './client';

export const login = (phone_number, password) =>
  api.post('/auth/login', { phone_number, password }).then((r) => r.data.data);

export const logout = () =>
  api.post('/auth/logout').then((r) => r.data);
