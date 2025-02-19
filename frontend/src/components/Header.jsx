import React from "react";
import { Link, useNavigate } from "react-router-dom";
import { getCurrentUser, logoutUser } from "../api/authApi";

const Header = () => {
  const navigate = useNavigate();
  const currentUser = getCurrentUser();

  const handleLogout = () => {
    logoutUser();
    navigate("/login");
  };

  return (
    <header className="header">
      <h1>Bad Jokes</h1>
      <div className="auth-info">
        {currentUser ? (
          <>
            <span>{currentUser.username}</span>
            <button className="auth-button" onClick={handleLogout}>Logout</button>
          </>
        ) : (
          <Link to="/login" className="auth-button">Login</Link>
        )}
      </div>
    </header>
  );
};

export default Header;
