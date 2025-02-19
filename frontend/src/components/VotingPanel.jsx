import React, { useState } from "react";
import { voteEntity } from "../api/jokesApi";
import { getCurrentUser } from "../api/authApi";

const VotingPanel = ({ entityType, entityId, initialScore, initialVote }) => {
  const currentUser = getCurrentUser();
  const userId = currentUser ? currentUser.userId : null;
  const isLoggedIn = !!currentUser;

  const [score, setScore] = useState(initialScore);
  const [hasVoted, setHasVoted] = useState(initialVote);

  const handleVote = async (voteType) => {
    if (!isLoggedIn) return;

    setScore((prevScore) => {
      if (hasVoted === voteType) {
        setHasVoted(null);
        return prevScore + (voteType === "plus" ? -1 : 1);
      } else {
        setHasVoted(voteType);
        return prevScore + (voteType === "plus" ? 1 : -1) * (hasVoted ? 2 : 1);
      }
    });

    await voteEntity(entityType, entityId, voteType);
  };

  return (
    <div className="rating">
      <button
        className="upvote"
        onClick={() => handleVote("plus")}
        disabled={!isLoggedIn}
        style={{
          color: hasVoted === "plus" ? "green" : "#888",
          cursor: isLoggedIn ? "pointer" : "not-allowed",
        }}
      >
        ▲
      </button>
      <div className="score">{score}</div>
      <button
        className="downvote"
        onClick={() => handleVote("minus")}
        disabled={!isLoggedIn}
        style={{
          color: hasVoted === "minus" ? "red" : "#888",
          cursor: isLoggedIn ? "pointer" : "not-allowed",
        }}
      >
        ▼
      </button>
    </div>
  );
};

export default VotingPanel;
