import React, { useState } from "react";
import ReactQuill from "react-quill-new";
import "react-quill-new/dist/quill.snow.css";
import { addComment } from "../api/commentsApi";

const CommentForm = ({ jokeId, parentId = null, onCommentAdded, isReply = false }) => {
    const [body, setBody] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false);

    const modules = {
        toolbar: [
            ["bold", "italic", "underline"],
            ["link", "blockquote", "code-block"],
            ["clean"]
        ]
    };

    const formats = [
        "bold", "italic", "underline",
        "link", "blockquote", "code-block"
    ];

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (body.trim() && !isSubmitting) {
            setIsSubmitting(true);
            try {
                const newComment = await addComment(jokeId, body, parentId);
                setBody("");
                if (onCommentAdded) {
                    onCommentAdded(newComment);
                }
            } catch (error) {
                console.error("Failed to add comment:", error);
            } finally {
                setIsSubmitting(false);
            }
        }
    };

    return (
        <div className={`comment-form ${isReply ? 'reply-form' : ''}`}>
            <h4>{isReply ? "Post a reply" : "Add a comment"}</h4>
            <div className="editor-container">
                <ReactQuill
                    theme="snow"
                    value={body}
                    onChange={setBody}
                    modules={modules}
                    formats={formats}
                    placeholder={isReply ? "Write your reply..." : "Write your comment..."}
                />
            </div>
            <div className="form-actions">
                {isReply && (
                    <button
                        className="cancel-button"
                        onClick={() => onCommentAdded()}
                        disabled={isSubmitting}
                    >
                        Cancel
                    </button>
                )}
                <button
                    className="submit-button"
                    onClick={handleSubmit}
                    disabled={!body.trim() || isSubmitting}
                >
                    {isSubmitting ? "Submitting..." : isReply ? "Reply" : "Comment"}
                </button>
            </div>
        </div>
    );
};

export default CommentForm;