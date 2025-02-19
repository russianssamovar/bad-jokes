import React, { useState, useRef, useEffect } from "react";
import { reactToEntity } from "../api/jokesApi";

const reactionMap = {
  laugh: "😂",
  heart: "❤️",
  neutral: "😐",
  surprised: "😲",
  fire: "🔥",
  poop: "💩",
  angry: "😡",
  monkey: "🙈",
  thumbs_up: "👍",
  thumbs_down: "👎",
};

const availableReactions = Object.keys(reactionMap);

const ReactionsList = ({ jokeId, initialReactions, initialUserReactions, isLoggedIn }) => {
  const [reactions, setReactions] = useState(initialReactions);
  const [userReactions, setUserReactions] = useState(new Set(initialUserReactions));
  const [showReactionPopup, setShowReactionPopup] = useState(false);
  const popupRef = useRef(null);

  useEffect(() => {
    setReactions(initialReactions);
    setUserReactions(new Set(initialUserReactions));
  }, [initialReactions, initialUserReactions]);

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (popupRef.current && !popupRef.current.contains(event.target)) {
        setShowReactionPopup(false);
      }
    };

    if (showReactionPopup) {
      document.addEventListener("mousedown", handleClickOutside);
    } else {
      document.removeEventListener("mousedown", handleClickOutside);
    }

    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [showReactionPopup]);

  const handleReactionClick = async (reaction) => {
    if (!isLoggedIn) return;

    setReactions((prevReactions) => {
      const updatedReactions = { ...prevReactions };
      const newUserReactions = new Set(userReactions);

      if (newUserReactions.has(reaction)) {
        updatedReactions[reaction] = Math.max((updatedReactions[reaction] || 1) - 1, 0);
        newUserReactions.delete(reaction);
        if (updatedReactions[reaction] === 0) {
          delete updatedReactions[reaction];
        }
      } else {
        updatedReactions[reaction] = (updatedReactions[reaction] || 0) + 1;
        newUserReactions.add(reaction);
      }

      setUserReactions(newUserReactions);
      return updatedReactions;
    });

    await reactToEntity("joke", jokeId, reaction);
    setShowReactionPopup(false);
  };

  return (
    <div className="reactions-container">
      <div className="reactions">
        {Object.entries(reactions).map(([reaction, count]) => (
          <div
            key={reaction}
            className="reaction"
            onClick={() => handleReactionClick(reaction)}
            style={{
              background: userReactions.has(reaction) ? "#d1e7fd" : "#f0f2f5",
              cursor: isLoggedIn ? "pointer" : "not-allowed",
            }}
          >
            {reactionMap[reaction]} <span>{count}</span>
          </div>
        ))}

        <div
          className="reaction add-reaction"
          onClick={isLoggedIn ? () => setShowReactionPopup(!showReactionPopup) : undefined}
          style={{ cursor: isLoggedIn ? "pointer" : "not-allowed" }}
        >
          ➕
        </div>
      </div>

      {showReactionPopup && (
        <div className="reaction-popup" ref={popupRef}>
          {availableReactions.map((reaction) => (
            <div
              key={reaction}
              className="reaction-option"
              onClick={() => handleReactionClick(reaction)}
            >
              {reactionMap[reaction]}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default ReactionsList;
