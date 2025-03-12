import { api } from '../utils/api';

export const fetchComments = async (jokeId) => {
  const response = await api.get(`/api/comments`, { params: { joke_id: jokeId } });
  return response.data;
};

export const addComment = async (jokeId, body, parentId = null) => {
  try {
    const response = await api.post(`/api/jokes/${jokeId}/comments`, {
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
    await api.delete(`/api/comments/${commentId}`);
    return true;
  } catch (error) {
    console.error("Error deleting comment:", error);
    throw error;
  }
};