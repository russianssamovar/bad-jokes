import axios from "axios";

const BASE_API_URL = import.meta.env.VITE_API_URL || "http://localhost:9999";
const API_URL = `${BASE_API_URL}/api/jokes`;

const apiClient = axios.create({
  baseURL: API_URL,
});

apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const fetchJokes = async ({ pageParam = 1, pageSize = 10, sortField = "created_at", order = "desc" }) => {
  const response = await apiClient.get("", { 
    params: { page: pageParam, pageSize, sortField, order }
  });
  return response.data;
};

export const voteEntity = async (entityType, entityId, voteType) => {
  return await apiClient.post(`/vote`, {
    entity_type: entityType,
    entity_id: entityId,
    vote_type: voteType,
  });
};

export const reactToEntity = async (entityType, entityId, reactionType) => {
  return await apiClient.post(`/react`, {
    entity_type: entityType,
    entity_id: entityId,
    reaction_type: reactionType,
  });
};

export const createJoke = async (body) => {
  const response = await apiClient.post("", { body });
  return response.data;
};

export const deleteJoke = async (jokeId) => {
  await apiClient.delete(`/delete`, {
    params: { joke_id: jokeId }
  });
};

export const addComment = async (jokeId, body) => {
  const response = await apiClient.post(`/comment`, {
    joke_id: jokeId,
    body
  });
  return response.data;
};

export const fetchComments = async (jokeId) => {
  const response = await apiClient.get(`/comments`, {
    params: { joke_id: jokeId }
  });
  return response.data;
};

export const fetchJokeWithComments = async (jokeId) => {
  try {
    const response = await apiClient.get(`/${jokeId}`);
    return response.data;
  } catch (error) {
    console.error("Error fetching joke with comments:", error);
    throw error;
  }
};