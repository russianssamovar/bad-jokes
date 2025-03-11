import axios from "axios";

const BASE_API_URL = import.meta.env.VITE_API_URL || "http://localhost:9999/api";
const API_URL = `${BASE_API_URL}/auth`;

export const registerUser = async (username, email, password) => {
  const response = await axios.post(`${API_URL}/register`, { username, email, password });
  return response.data;
};

export const loginUser = async (email, password) => {
  const response = await axios.post(`${API_URL}/login`, { email, password });
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
