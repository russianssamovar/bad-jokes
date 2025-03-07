import React, { useState, useEffect } from "react";
import { useNavigate, Link } from "react-router-dom";
import { createJoke } from "../api/jokesApi";
import { getCurrentUser } from "../api/authApi";
import ReactQuill from "react-quill-new";
import "react-quill-new/dist/quill.snow.css";
import JokeCard from "./JokeCard";

const CreateJoke = () => {
  const [body, setBody] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showPreview, setShowPreview] = useState(false);
  const navigate = useNavigate();
  const user = getCurrentUser();

  useEffect(() => {
    if (!user) {
      navigate("/auth");
    }
  }, [user, navigate]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (body.trim() && !isSubmitting) {
      setIsSubmitting(true);
      try {
        await createJoke(body);
        navigate("/");
      } catch (error) {
        console.error("Failed to create joke:", error);
      } finally {
        setIsSubmitting(false);
      }
    }
  };

  const togglePreview = () => {
    setShowPreview(!showPreview);
  };

  const isContentEmpty = (html) => {
    const tempDiv = document.createElement("div");
    tempDiv.innerHTML = html;
    return tempDiv.textContent.trim().length === 0;
  };

  const modules = {
    toolbar: [
      ["bold", "italic", "underline", "strike"],
      [{ header: 1 }, { header: 2 }],
      [{ list: "ordered" }],
      ["link", "blockquote", "code-block", "image"],
      [{ color: [] }, { background: [] }],
      [{ align: [] }],
      ["clean"]
    ]
  };

  const formats = [
    "bold", "italic", "underline", "strike",
    "header",
    "list",
    "link", "blockquote", "code-block", "image",
    "color", "background",
    "align"
  ];

  const previewJoke = {
    id: "preview",
    body: body,
    author_id: user?.userId,
    author_username: user?.username,
    comment_count: 0,
    social: {
      pluses: 0,
      minuses: 0,
      reactions: [],
      user: { reactions: [], vote_type: null }
    }
  };

  return (
      <div className="create-joke-container">
        <div className="navigation-section">
          <Link to="/" className="back-link">
            &larr; Back to jokes
          </Link>
        </div>

        <div className="create-joke-card">
          <h2>{showPreview ? "Preview Your Joke" : "Create a New Joke"}</h2>

          {!showPreview ? (
              <div className="editor-container">
                <ReactQuill
                    theme="snow"
                    value={body}
                    onChange={setBody}
                    modules={modules}
                    formats={formats}
                    placeholder="Write your joke here..."
                />
              </div>
          ) : (
              <div className="joke-list-container">
                <JokeCard joke={previewJoke} />
              </div>
          )}

          <div className="form-actions">
            <button
                className={`sort-button ${showPreview ? 'active' : ''}`}
                onClick={togglePreview}
                disabled={isContentEmpty(body)}
            >
              {showPreview ? "Edit" : "Preview"}
            </button>
            <button
                className="submit-button"
                onClick={handleSubmit}
                disabled={isContentEmpty(body) || isSubmitting}
            >
              {isSubmitting ? "Posting..." : "Post Joke"}
            </button>
          </div>
        </div>
      </div>
  );
};

export default CreateJoke;