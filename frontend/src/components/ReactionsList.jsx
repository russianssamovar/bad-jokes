import React, { useState, useRef, useEffect } from "react";
import { reactToEntity } from "../api/jokesApi";

const reactionMap = {
  laugh: "😂", heart: "❤️", neutral: "😐", surprised: "😲", fire: "🔥",
  poop: "💩", angry: "😡", monkey: "🙈", thumbs_up: "👍", thumbs_down: "👎",
};

const availableReactions = Object.keys(reactionMap);

const ReactionsList = ({ jokeId, initialReactions, initialUserReactions, isLoggedIn }) => {
  const [reactions, setReactions] = useState({ ...initialReactions });
  const [userReactions, setUserReactions] = useState(new Set(initialUserReactions || []));
  const [showReactionPopup, setShowReactionPopup] = useState(false);
  const [splashEffect, setSplashEffect] = useState(null);
  const popupRef = useRef(null);
  const addButtonRef = useRef(null);
  const reactionsRef = useRef({});

  useEffect(() => {
    setReactions({ ...initialReactions });
    setUserReactions(new Set(initialUserReactions || []));
  }, [initialReactions, initialUserReactions]);

  useEffect(() => {
    if (!showReactionPopup) return;

    const handleClickOutside = (event) => {
      if (popupRef.current && !popupRef.current.contains(event.target) &&
          addButtonRef.current && !addButtonRef.current.contains(event.target)) {
        setShowReactionPopup(false);
      }
    };

    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [showReactionPopup]);

  const handleReactionClick = async (reaction, event) => {
    if (!isLoggedIn) return;

    const isAddingReaction = !userReactions.has(reaction);

    setSplashEffect({
      reaction,
      isAdding: isAddingReaction,
    });

    setTimeout(() => setSplashEffect(null), 600);

    setReactions(prevReactions => {
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
                      ref={el => reactionsRef.current[reaction] = el}
                  >
                    {reactionMap[reaction]} <span>{count}</span>
                  </div>
              ))}

          {splashEffect && (
              <div
                  className="reaction-splash-container"
                  style={{
                    position: "absolute",
                    left: `${addButtonRef.current.closest(".reactions-wrapper").getBoundingClientRect().width - 70}px`,
                    top: `${addButtonRef.current.getBoundingClientRect().top.toFixed()}`,
                  }}
              >
                {[...Array(8)].map((_, i) => (
                    <span
                        key={i}
                        className={`reaction-splash ${!splashEffect.isAdding ? "removal" : ""}`}
                        data-reaction={splashEffect.reaction}
                    >
                {reactionMap[splashEffect.reaction]}
              </span>
                ))}
              </div>
          )}

          <div
              className="reaction add-reaction"
              ref={addButtonRef}
              onClick={() => isLoggedIn && setShowReactionPopup(!showReactionPopup)}
          >
            ➕
          </div>
        </div>

        {showReactionPopup && addButtonRef.current && (
            <div
                className="reaction-popup"
                ref={popupRef}
                style={{
                  position: "absolute",
                  left: `${addButtonRef.current.closest(".reactions-wrapper").getBoundingClientRect().width}px`,
                  top: addButtonRef.current.getBoundingClientRect().bottom.toFixed(),
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
      </div>
  );
};

export default ReactionsList;