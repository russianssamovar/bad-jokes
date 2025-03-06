import axios from "axios";

const BASE_API_URL = import.meta.env.VITE_API_URL || "http://localhost:9999";
const API_URL = `${BASE_API_URL}/api/comments`;

export const fetchComments = async (jokeId) => {
  const response = await axios.get(`${API_URL}`, { params: { joke_id: jokeId } });
  return response.data;
};

export const addComment = async (jokeId, body, parentCommentId = null) => {
  await axios.post(`${API_URL}`, { joke_id: jokeId, body, parent_comment_id: parentCommentId });
}; 