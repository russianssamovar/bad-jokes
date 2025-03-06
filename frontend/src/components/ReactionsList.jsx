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
  const [reactions, setReactions] = useState({ ...initialReactions });
  const [userReactions, setUserReactions] = useState(new Set(initialUserReactions || []));
  const [showReactionPopup, setShowReactionPopup] = useState(false);
  const [reactionEffect, setReactionEffect] = useState(null);
  const [removalEffect, setRemovalEffect] = useState(null); 
  const popupRef = useRef(null);
  const buttonRef = useRef(null);

  useEffect(() => {
    setReactions({ ...initialReactions });
    setUserReactions(new Set(initialUserReactions || []));
  }, [initialReactions, initialUserReactions]);

  useEffect(() => {
    const handleClickOutside = (event) => {
      if (
        popupRef.current &&
        !popupRef.current.contains(event.target) &&
        buttonRef.current &&
        !buttonRef.current.contains(event.target)
      ) {
        setShowReactionPopup(false);
      }
    };

    if (showReactionPopup) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [showReactionPopup]);

  const handleReactionClick = async (reaction, event) => {
    if (!isLoggedIn) return;

    const rect = event.currentTarget.getBoundingClientRect();
    const effectPosition = { x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 };

    if (userReactions.has(reaction)) {
      setRemovalEffect({ reaction, ...effectPosition });
      setTimeout(() => setRemovalEffect(null), 600);
    } else {
      setReactionEffect({ reaction, ...effectPosition });
      setTimeout(() => setReactionEffect(null), 600);
    }

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
    <div className="reactions-wrapper">
      <div className="reactions">
        {Object.entries(reactions)
          .filter(([_, count]) => count > 0)
          .map(([reaction, count]) => (
            <div
              key={reaction}
              className={`reaction ${userReactions.has(reaction) ? "active" : ""}`}
              onClick={(e) => handleReactionClick(reaction, e)}
            >
              {reactionMap[reaction]} <span>{count}</span>
            </div>
          ))}
          
        <div
          className="reaction add-reaction"
          ref={buttonRef}
          onClick={() => isLoggedIn && setShowReactionPopup(!showReactionPopup)}
        >
          ➕
        </div>
      </div>
      
      {showReactionPopup && buttonRef.current && (
        <div
          className="reaction-popup"
          ref={popupRef}
          style={{
            left: `${buttonRef.current.closest(".reactions-wrapper").getBoundingClientRect().width}px`,
            top: buttonRef.current.getBoundingClientRect().bottom.toFixed(),
          }}
        >
          {availableReactions.map((reaction) => (
            <div
              key={reaction}
              className={`reaction-option ${userReactions.has(reaction) ? "active" : ""}`}
              onClick={(e) => handleReactionClick(reaction, e)}
            >
              {reactionMap[reaction]}
            </div>
          ))}
        </div>
      )}

      {reactionEffect && (
        <div className="reaction-splash-container" style={{ left: reactionEffect.x, top: reactionEffect.y }}>
          {[...Array(8)].map((_, i) => (
            <span key={i} className="reaction-splash" data-reaction={reactionEffect.reaction}>
              {reactionMap[reactionEffect.reaction]}
            </span>
          ))}
        </div>
      )}

      {removalEffect && (
        <div className="reaction-splash-container" style={{ left: removalEffect.x, top: removalEffect.y }}>
          {[...Array(8)].map((_, i) => (
            <span key={i} className="reaction-splash removal" data-reaction={removalEffect.reaction}>
              {reactionMap[removalEffect.reaction]}
            </span>
          ))}
        </div>
      )}
    </div>
  );
};

export default ReactionsList;
