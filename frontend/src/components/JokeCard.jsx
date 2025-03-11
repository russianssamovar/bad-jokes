import React, { useState } from "react";
import { deleteJoke } from "../api/jokesApi";
import { deleteAsAdminJoke } from "../api/adminApi";
import { getCurrentUser } from "../api/authApi";
import ReactionsList from "./ReactionsList";
import VotingPanel from "./VotingPanel";
import { Link } from "react-router-dom";
import { formatDistanceToNow } from "date-fns";
import Popup from "./Popup";

const JokeCard = ({ joke, onDelete }) => {
    const currentUser = getCurrentUser();
    const isAuthor = currentUser?.userId === joke.author_id;
    const isAdmin = currentUser?.isAdmin;
    const [showDeletePopup, setShowDeletePopup] = useState(false);
    
    const getInitials = (username) => {
        return username ? username.charAt(0).toUpperCase() : '?';
    };

    const handleDelete = async () => {
        setShowDeletePopup(true);
    };

    const confirmDelete = async () => {
        if (isAdmin){
            await deleteAsAdminJoke(joke.id);
        }
        else {
            await deleteJoke(joke.id);
        }
        
        if (onDelete) onDelete(joke.id);
        setShowDeletePopup(false);
    };
    
    return (
        <>
            <div className="joke-card">
                <div className="joke-header">
                    <div className="user-info">
                        <div className="user-avatar-small">
                            {getInitials(joke.author_username)}
                        </div>
                        <span className="comment-author">{joke.author_username}</span>
                    </div>
                    <div className="joke-meta">
                        <span className="comment-time">
                            {joke.created_at && formatDistanceToNow(new Date(joke.created_at))} ago
                        </span>
                        {(isAuthor || isAdmin) && (
                            <button className="delete-button" onClick={handleDelete}>
                                <svg width="18" height="18" viewBox="0 0 24 24">
                                    <line x1="4" y1="4" x2="20" y2="20" stroke="black" strokeWidth="2"/>
                                    <line x1="20" y1="4" x2="4" y2="20" stroke="black" strokeWidth="2"/>
                                </svg>
                            </button>
                        )}
                    </div>
                </div>

                <div className="joke-content-row">
                    <div
                        className="joke-text rich-content"
                        dangerouslySetInnerHTML={{__html: joke.body}}
                    />
                </div>

                <ReactionsList
                    entityType="joke"
                    entityId={joke.id}
                    initialReactions={joke.social.reactions}
                    initialUserReactions={joke.social?.user?.reactions}
                    isLoggedIn={!!currentUser}
                />

                <div className="bottom-panel">
                    <Link to={`/joke/${joke.id}`} className="comment-count">
                        💬 {joke.comment_count} comments
                    </Link>
                    <VotingPanel
                        entityType="joke"
                        entityId={joke.id}
                        initialScore={joke.social.pluses - joke.social.minuses}
                        initialVote={joke.social?.user?.vote_type}
                    />
                </div>
            </div>
            
            <Popup
                isOpen={showDeletePopup}
                title="Delete Joke"
                message="Are you sure you want to delete this joke? This action cannot be undone."
                onConfirm={confirmDelete}
                onCancel={() => setShowDeletePopup(false)}
                confirmText="Delete"
                cancelText="Cancel"
                type="delete"
            />
        </>
    );
};

export default JokeCard;