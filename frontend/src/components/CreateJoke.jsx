import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { createJoke } from "../api/jokesApi";
import { getCurrentUser } from "../api/authApi";

const CreateJoke = () => {
  const [body, setBody] = useState("");
  const navigate = useNavigate();
  const user = getCurrentUser();

  useEffect(() => {
    if (!user) {
      navigate("/login");
    }
  }, [user, navigate]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (body.trim()) {
      await createJoke(body);
      navigate("/");
    }
  };

  return (
    <div className="auth-form">
      <h2>Create a New Joke</h2>
      <textarea
        rows="5"
        value={body}
        onChange={(e) => setBody(e.target.value)}
        placeholder="Write your joke here..."
      ></textarea>
      <button onClick={handleSubmit}>Post</button>
    </div>
  );
};

export default CreateJoke;
