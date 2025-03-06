import React, { useState } from "react";
import { voteEntity } from "../api/jokesApi";
import { getCurrentUser } from "../api/authApi";

const VotingPanel = ({ entityType, entityId, initialScore, initialVote }) => {
  const currentUser = getCurrentUser();
  const isLoggedIn = !!currentUser;

  const [score, setScore] = useState(initialScore);
  const [hasVoted, setHasVoted] = useState(initialVote || null);
  const [showEffect, setShowEffect] = useState(null);

  const handleVote = async (voteType) => {
    if (!isLoggedIn) return;

    setShowEffect(voteType);
    setTimeout(() => setShowEffect(null), 500);

    const currentVote = hasVoted;
    const currentScore = score;

    let newVote;
    let newScore;

    if (currentVote === voteType) {
      newVote = null;
      newScore = currentScore + (voteType === "plus" ? -1 : 1);
    } else if (currentVote === null) {
      newVote = voteType;
      newScore = currentScore + (voteType === "plus" ? 1 : -1);
    } else {
      newVote = voteType;
      newScore = currentScore + (voteType === "plus" ? 2 : -2);
    }

    setHasVoted(newVote);
    setScore(newScore);

    await voteEntity(entityType, entityId, voteType);
  };

  return (
    <div className="voting-panel">
      <button className={`voting-button upvote ${hasVoted === "plus" ? "active-upvote" : ""}`} onClick={() => handleVote("plus")} disabled={!isLoggedIn}>
        <svg viewBox="0 0 24 24">
          <polyline points="6 15 12 9 18 15" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
        {showEffect === "plus" && (
          <div className="vote-splash-container">
            {[...Array(6)].map((_, i) => (
              <span key={i} className="vote-splash upvote-splash"></span>
            ))}
          </div>
        )}
      </button>

      <div className="voting-score">{score}</div>

      <button className={`voting-button downvote ${hasVoted === "minus" ? "active-downvote" : ""}`} onClick={() => handleVote("minus")} disabled={!isLoggedIn}>
        <svg viewBox="0 0 24 24">
          <polyline points="6 9 12 15 18 9" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"/>
        </svg>
        {showEffect === "minus" && (
          <div className="vote-splash-container">
            {[...Array(6)].map((_, i) => (
              <span key={i} className="vote-splash downvote-splash"></span>
            ))}
          </div>
        )}
      </button>
    </div>
  );
};

export default VotingPanel;
