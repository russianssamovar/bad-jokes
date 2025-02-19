import React, { useState, useEffect } from "react";
import { fetchComments, addComment } from "../api/commentsApi";

const Comments = ({ jokeId }) => {
  const [comments, setComments] = useState([]);
  const [newComment, setNewComment] = useState("");

  useEffect(() => {
    const loadComments = async () => {
      const fetchedComments = await fetchComments(jokeId);
      setComments(fetchedComments);
    };
    loadComments();
  }, [jokeId]);

  const handleAddComment = async () => {
    if (newComment.trim()) {
      await addComment(jokeId, newComment);
      setNewComment("");
      const updatedComments = await fetchComments(jokeId);
      setComments(updatedComments);
    }
  };

  return (
    <div>
      <h3>Comments</h3>
      <ul>
        {comments.map((comment) => (
          <li key={comment.id}>{comment.body}</li>
        ))}
      </ul>
      <textarea
        value={newComment}
        onChange={(e) => setNewComment(e.target.value)}
        placeholder="Add a comment..."
      />
      <button onClick={handleAddComment}>Submit</button>
    </div>
  );
};

export default Comments; 