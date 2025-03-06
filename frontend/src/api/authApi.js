import axios from "axios";

const BASE_API_URL = import.meta.env.VITE_API_URL || "http://localhost:9999";
const API_URL = `${BASE_API_URL}/api/auth`;

export const registerUser = async (username, email, password) => {
  const response = await axios.post(`${API_URL}/register`, { username, email, password });
  return response.data;
};

export const loginUser = async (email, password) => {
  const response = await axios.post(`${API_URL}/login`, { email, password });
  return response.data; 
};

export const getCurrentUser = () => {
  const token = localStorage.getItem("token");
  if (!token) return null;

  const payload = JSON.parse(atob(token.split('.')[1]));
  return { userId: payload.user_id, username: payload.username, token };
};


export const logoutUser = () => {
  localStorage.removeItem("token");
};
