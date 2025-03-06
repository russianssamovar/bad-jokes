import React, { useState, useEffect } from "react";
import { useParams, Link } from "react-router-dom";
import { fetchJokeWithComments } from "../api/jokesApi";
import JokeCard from "../components/JokeCard";
import CommentList from "../components/CommentList";
import CommentForm from "../components/CommentForm";
import { getCurrentUser } from "../api/authApi";

const JokeDetail = () => {
    const { jokeId } = useParams();
    const [joke, setJoke] = useState(null);
    const [comments, setComments] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const currentUser = getCurrentUser();

    useEffect(() => {
        const loadJokeWithComments = async () => {
            try {
                setLoading(true);
                const data = await fetchJokeWithComments(jokeId);
                setJoke(data.joke);
                setComments(data.comments);
                setError(null);
            } catch (err) {
                setError("Failed to load joke and comments");
                console.error(err);
            } finally {
                setLoading(false);
            }
        };

        loadJokeWithComments();
    }, [jokeId]);

    const handleCommentAdded = (newComment) => {
        setComments(prevComments => [...(prevComments || []), newComment]);
    };

    const handleCommentDeleted = (commentId) => {
        setComments((prevComments) =>
            prevComments.filter((comment) => comment.id !== commentId)
        );
    };

    return (
        <div className="joke-detail-container">
            <Link to="/" className="back-link">
                &larr; Back to jokes
            </Link>

            {loading && <div className="loading">Loading...</div>}

            {error && <div className="error-message">{error}</div>}

            {joke && (
                <>
                    <div className="joke-section">
                        <JokeCard joke={joke} showLink={false} />
                    </div>

                    <div className="comments-section">
                        <h3 className="comments-header">Comments ({joke.comment_count})</h3>

                        {currentUser && (
                            <CommentForm
                                jokeId={jokeId}
                                onCommentAdded={handleCommentAdded}
                            />
                        )}

                        {comments?.length > 0 ? (
                            <CommentList
                                comments={comments}
                                onCommentDeleted={handleCommentDeleted}
                            />
                        ) : (
                            <div className="no-comments">No comments yet. Be the first to comment!</div>
                        )}
                    </div>
                </>
            )}
        </div>
    );
};

export default JokeDetail;