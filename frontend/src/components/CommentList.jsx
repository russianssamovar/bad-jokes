import React from "react";
import Comment from "./Comment";

const CommentList = ({ comments, onCommentDeleted, onReplyAdded }) => {
    const commentMap = {};
    const rootComments = [];

    comments.forEach(comment => {
        commentMap[comment.id] = {
            ...comment,
            children: []
        };
    });

    comments.forEach(comment => {
        if (comment.parent_id) {
            if (commentMap[comment.parent_id]) {
                commentMap[comment.parent_id].children.push(commentMap[comment.id]);
            } else {
                rootComments.push(commentMap[comment.id]);
            }
        } else {
            rootComments.push(commentMap[comment.id]);
        }
    });

    return (
        <div className="comment-list">
            {rootComments.map(comment => (
                <Comment
                    key={comment.id}
                    comment={comment}
                    onCommentDeleted={onCommentDeleted}
                    onReplyAdded={onReplyAdded}
                />
            ))}
        </div>
    );
};

export default CommentList;