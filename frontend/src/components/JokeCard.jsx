import React from "react";
import { deleteJoke } from "../api/jokesApi";
import { getCurrentUser } from "../api/authApi";
import ReactionsList from "./ReactionsList";
import VotingPanel from "./VotingPanel";

const JokeCard = ({ joke, onDelete }) => {
  const currentUser = getCurrentUser();
  const userId = currentUser ? currentUser.userId : null;
  const isLoggedIn = !!currentUser;

  const isAuthor = joke.author_id === userId;

  const handleDelete = async () => {
    if (window.confirm("Are you sure you want to delete this joke?")) {
      await deleteJoke(joke.id);
      if (onDelete) onDelete(joke.id);
    }
  };

  return (
    <div className="joke-card">
      <VotingPanel
        entityType="joke"
        entityId={joke.id}
        initialScore={joke.social.pluses - joke.social.minuses}
        initialVote={joke.social.user?.vote_type || null}
      />

      <div className="joke-content">
        <p>{joke.body}</p>

        <ReactionsList
          jokeId={joke.id}
          initialReactions={joke.social.reactions}
          initialUserReactions={joke.social.user?.reactions || []}
          isLoggedIn={isLoggedIn}
        />

        <div className="comment-count">💬 {joke.comment_count} comments</div>

        {isAuthor && <button className="delete-button" onClick={handleDelete}>🗑️</button>}
      </div>
    </div>
  );
};

export default JokeCard;
