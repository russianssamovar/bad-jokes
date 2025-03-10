import axios from "axios";

const BASE_API_URL = import.meta.env.VITE_API_URL || `${window.location.protocol}//${window.location.host}`;

const apiClient = axios.create({
  baseURL: BASE_API_URL,
});

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const fetchComments = async (jokeId) => {
  const response = await apiClient.get(`/api/comments`, { params: { joke_id: jokeId } });
  return response.data;
};

export const addComment = async (jokeId, body, parentId = null) => {
  try {
    const response = await apiClient.post(`/api/jokes/${jokeId}/comments`, {
      body,
      parent_id: parentId
    });
    return response.data;
  } catch (error) {
    console.error("Error adding comment:", error);
    throw error;
  }
};

export const deleteComment = async (commentId) => {
  try {
    await apiClient.delete(`/api/comments/${commentId}`);
    return true;
  } catch (error) {
    console.error("Error deleting comment:", error);
    throw error;
  }
};