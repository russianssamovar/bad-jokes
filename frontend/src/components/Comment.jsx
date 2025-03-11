import React, { useState } from "react";
import { formatDistanceToNow } from "date-fns";
import { deleteComment } from "../api/commentsApi";
import { getCurrentUser } from "../api/authApi";
import VotingPanel from "./VotingPanel";
import ReactionsList from "./ReactionsList";
import CommentForm from "./CommentForm";
import Popup from "./Popup";
import {deleteAsAdminComment} from "../api/adminApi.js";

const Comment = ({ comment, onCommentDeleted, onReplyAdded }) => {
    const [showReplyForm, setShowReplyForm] = useState(false);
    const [showDeletePopup, setShowDeletePopup] = useState(false);
    const currentUser = getCurrentUser();
    const isAuthor = currentUser?.userId === comment.author_id;
    const isAdmin = currentUser?.isAdmin;

    const handleDelete = () => {
        setShowDeletePopup(true);
    };
    
    const confirmDelete = async () => {
        try {
            if (isAdmin) {
                await deleteAsAdminComment(comment.id);
            }
            else {
                await deleteComment(comment.id);
            }
            onCommentDeleted(comment.id);
            setShowDeletePopup(false);
        } catch (error) {
            console.error("Failed to delete comment:", error);
        }
    };
    
    const toggleReplyForm = () => {
        setShowReplyForm(!showReplyForm);
    };

    const handleReplyAdded = (newReply) => {
        setShowReplyForm(false);

        if (newReply && onReplyAdded) {
            onReplyAdded(newReply);
        }
    };

    return (
        <>
            <div className={`comment ${comment.is_deleted ? 'comment-deleted' : ''}`}>
                <div className="comment-header">
                    <span className="comment-author">{comment.author_username}</span>
                    <span className="comment-time">
                  {formatDistanceToNow(new Date(comment.created_at))} ago
                </span>
                </div>

                {comment.is_deleted ? (
                    <div className="comment-body deleted-comment">
                        <i>This comment has been deleted</i>
                    </div>
                ) : (
                    <div
                        className="comment-body rich-content"
                        dangerouslySetInnerHTML={{__html: comment.body}}
                    />
                )}

                {!comment.is_deleted && (
                    <div className="comment-actions">
                        <ReactionsList
                            entityType="comment"
                            entityId={comment.id}
                            initialReactions={comment.social.reactions}
                            initialUserReactions={comment.social?.user?.reactions}
                            isLoggedIn={!!currentUser}
                        />

                        <div className="comment-buttons">
                            {currentUser && (
                                <button className="reply-button" onClick={toggleReplyForm}>
                                    Reply
                                </button>
                            )}

                            {(isAuthor || isAdmin) && (
                                <button className="delete-button" onClick={handleDelete}>
                                    Delete
                                </button>
                            )}
                        </div>

                        <VotingPanel
                            entityType="comment"
                            entityId={comment.id}
                            initialScore={comment.social.pluses}
                            initialVote={comment.social?.user?.vote_type}
                        />
                    </div>
                )}

                {showReplyForm && currentUser && (
                    <div className="reply-form-container">
                        <CommentForm
                            jokeId={comment.joke_id}
                            parentId={comment.id}
                            onCommentAdded={handleReplyAdded}
                            isReply
                        />
                    </div>
                )}

                {comment.children && comment.children.length > 0 && (
                    <div className="nested-comments">
                        {comment.children.map(childComment => (
                            <Comment
                                key={childComment.id}
                                comment={childComment}
                                onCommentDeleted={onCommentDeleted}
                                onReplyAdded={onReplyAdded}
                            />
                        ))}
                    </div>
                )}
            </div>
            
            <Popup
                isOpen={showDeletePopup}
                title="Delete Comment"
                message="Are you sure you want to delete this comment? This action cannot be undone."
                onConfirm={confirmDelete}
                onCancel={() => setShowDeletePopup(false)}
                confirmText="Delete"
                cancelText="Cancel"
                type="delete"
            />
        </>
    );
};

export default Comment;