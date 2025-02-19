import React, { useState, useEffect } from "react";
import { useParams } from "react-router-dom";
import { fetchComments } from "../api/commentsApi";
import JokeBody from "../components/JokeBody";
import VoteElement from "../components/VoteElement";
import ReactionElement from "../components/ReactionElement";

const CommentsPage = () => {
  const { jokeId } = useParams();
  const [comments, setComments] = useState([]);
  const [joke, setJoke] = useState(null);

  useEffect(() => {
    const loadComments = async () => {
      const fetchedComments = await fetchComments(jokeId);
      setComments(fetchedComments);
    };
    loadComments();
  }, [jokeId]);

  return (
    <div>
      {joke && (
        <>
          <JokeBody joke={joke} />
          <VoteElement joke={joke} />
          <ReactionElement joke={joke} />
        </>
      )}
      <h3>Comments</h3>
      <ul>
        {comments.map((comment) => (
          <li key={comment.id}>
            <p>{comment.body}</p>
            {comment.is_author && <span>(You)</span>}
            <VoteElement comment={comment} />
            <ReactionElement comment={comment} />
          </li>
        ))}
      </ul>
    </div>
  );
};

export default CommentsPage; 