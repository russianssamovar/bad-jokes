import { api } from '../utils/api';
export const registerUser = async (username, email, password) => {
  const response = await api.post(`/auth/register`, { username, email, password });
  return response.data;
};

export const loginUser = async (email, password) => {
  const response = await api.post(`/auth/login`, { email, password });
  return response.data;
};

export const getCurrentUser = () => {
  try {
    const token = localStorage.getItem("token");
    if (!token) return null;

    const parts = token.split('.');
    if (parts.length !== 3) return null;

    const payload = JSON.parse(atob(parts[1]));
    if (!payload || !payload.user_id || !payload.username) return null;

    return { userId: payload.user_id, username: payload.username, isAdmin: payload.is_admin, token };
  } catch (error) {
    console.error("Error parsing user token:", error);
    localStorage.removeItem("token");
    return null;
  }
};

export const logoutUser = () => {
  localStorage.removeItem("token");
};

export const handleOAuthCallback = async () => {
  const urlParams = new URLSearchParams(window.location.search);
  const token = urlParams.get('token');

  if (token) {
    localStorage.setItem('token', token);
    return getCurrentUser();
  }
  return null;
};