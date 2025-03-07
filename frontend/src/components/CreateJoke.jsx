import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { createJoke } from "../api/jokesApi";
import { getCurrentUser } from "../api/authApi";
import ReactQuill from "react-quill-new";
import "react-quill-new/dist/quill.snow.css";

const CreateJoke = () => {
  const [body, setBody] = useState("");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const navigate = useNavigate();
  const user = getCurrentUser();

  useEffect(() => {
    if (!user) {
      navigate("/login");
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

  return (
      <div className="create-joke-container">
        <div className="create-joke-card">
          <h2>Create a New Joke</h2>
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
          <button
              className="submit-button"
              onClick={handleSubmit}
              disabled={!body.trim() || isSubmitting}
          >
            {isSubmitting ? "Posting..." : "Post Joke"}
          </button>
        </div>
      </div>
  );
};

export default CreateJoke;