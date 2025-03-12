import { api } from '../utils/api';

export const fetchJokes = async ({ pageParam = 1, pageSize = 10, sortField = "created_at", order = "desc" }) => {
  const response = await api.get("/jokes", {
    params: { page: pageParam, page_size: pageSize, sort_field: sortField, order }
  });
  return response.data;
};

export const voteEntity = async (entityType, entityId, voteType) => {
  return await api.post(`/jokes/vote`, {
    entity_type: entityType,
    entity_id: entityId,
    vote_type: voteType,
  });
};

export const reactToEntity = async (entityType, entityId, reactionType) => {
  return await api.post(`/jokes/react`, {
    entity_type: entityType,
    entity_id: entityId,
    reaction_type: reactionType,
  });
};

export const createJoke = async (body) => {
  const response = await api.post("/jokes", { body });
  return response.data;
};

export const deleteJoke = async (jokeId) => {
  await api.delete(`/jokes/${jokeId}`);
};

export const addComment = async (jokeId, body) => {
  const response = await api.post(`/jokes/comment`, {
    joke_id: jokeId,
    body
  });
  return response.data;
};

export const fetchComments = async (jokeId) => {
  const response = await api.get(`/jokes/${jokeId}/comments`);
  return response.data;
};

export const fetchJokeWithComments = async (jokeId) => {
  try {
    const response = await api.get(`/jokes/${jokeId}`);
    return response.data;
  } catch (error) {
    console.error("Error fetching joke with comments:", error);
    throw error;
  }
};