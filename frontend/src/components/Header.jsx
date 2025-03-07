import React from "react";
import { Link, useNavigate } from "react-router-dom";
import { getCurrentUser, logoutUser } from "../api/authApi";

const Header = () => {
    const navigate = useNavigate();
    const currentUser = getCurrentUser();

    const getInitials = (username) => {
        return username ? username.charAt(0).toUpperCase() : '?';
    };

    const handleLogout = () => {
        logoutUser();
        navigate("/auth");
    };

    return (
        <header className="header">
            <div className="header-container">
                <Link to="/" className="logo-link">
                    <div className="logo-container">
                        <svg className="logo-svg" width="120" height="35" viewBox="0 0 120 35" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M25 5C15.5 5 8 12.61 8 22C8 24.96 8.81 27.73 10.27 30.06L8 35L13.16 32.36C16.71 34.33 20.69 35 25 35C34.5 35 42 27.39 42 18C42 8.61 34.5 5 25 5Z" fill="var(--primary-light)" stroke="var(--primary)" strokeWidth="2"/>
                            <path d="M17 17C18.1 17 19 16.1 19 15C19 13.9 18.1 13 17 13C15.9 13 15 13.9 15 15C15 16.1 15.9 17 17 17Z" fill="var(--primary)"/>
                            <path d="M33 17C34.1 17 35 16.1 35 15C35 13.9 34.1 13 33 13C31.9 13 31 13.9 31 15C31 16.1 31.9 17 33 17Z" fill="var(--primary)"/>
                            <path d="M17 29H33C33 25 29.5 22 25 22C20.5 22 17 25 17 29Z" stroke="var(--primary)" strokeWidth="2" strokeLinecap="round"/>
                            <path d="M52 10H57.1C59.3 10 61 10.4 62.1 11.3C63.2 12.1 63.8 13.3 63.8 14.9C63.8 16.5 63.2 17.7 62.1 18.6C61 19.5 59.3 20 57.1 20H52V10ZM57 17.9C58.3 17.9 59.3 17.6 59.9 17.1C60.5 16.6 60.8 15.8 60.8 14.9C60.8 14 60.5 13.3 59.9 12.8C59.3 12.3 58.3 12.1 57 12.1H54.9V17.9H57Z" fill="var(--text-dark)"/>
                            <path d="M70.5 20.2C69.3 20.2 68.3 19.9 67.4 19.4C66.5 18.8 65.9 18 65.4 17.1C64.9 16.1 64.7 15 64.7 13.8C64.7 12.6 65 11.5 65.5 10.6C66 9.7 66.7 8.9 67.6 8.3C68.5 7.7 69.5 7.4 70.7 7.4C71.9 7.4 72.9 7.7 73.8 8.3C74.7 8.9 75.3 9.7 75.8 10.6C76.3 11.5 76.5 12.6 76.5 13.8C76.5 15 76.2 16.1 75.7 17C75.2 17.9 74.4 18.7 73.5 19.3C72.5 19.9 71.6 20.2 70.5 20.2ZM70.5 17.9C71.1 17.9 71.6 17.8 72.1 17.5C72.5 17.2 72.9 16.7 73.1 16.2C73.3 15.6 73.5 14.8 73.5 13.9C73.5 13 73.4 12.2 73.1 11.6C72.9 11 72.6 10.6 72.1 10.3C71.7 10 71.2 9.8 70.6 9.8C70 9.8 69.4 10 69 10.3C68.5 10.6 68.2 11 67.9 11.6C67.7 12.2 67.6 12.9 67.6 13.9C67.6 14.8 67.7 15.6 68 16.2C68.2 16.8 68.5 17.2 69 17.5C69.4 17.8 69.9 17.9 70.5 17.9Z" fill="var(--text-dark)"/>
                            <path d="M87.9 10V20H85.2V10H87.9ZM92.5 10L88.4 14.8L84.5 10H82L87.2 16.2V20H89.9V16.2L95 10H92.5Z" fill="var(--text-dark)"/>
                            <path d="M101.8 20.2C100.6 20.2 99.5 19.9 98.6 19.4C97.7 18.8 97 18 96.5 17C96 16 95.8 14.9 95.8 13.7C95.8 12.5 96 11.4 96.5 10.4C97 9.4 97.7 8.6 98.6 8.1C99.5 7.5 100.6 7.2 101.8 7.2C103 7.2 104 7.5 104.9 8C105.8 8.6 106.5 9.3 106.9 10.3L104.7 11.3C104.4 10.7 104 10.2 103.5 9.9C103 9.6 102.4 9.4 101.8 9.4C101.1 9.4 100.5 9.6 100 9.9C99.5 10.2 99.1 10.7 98.8 11.3C98.5 11.9 98.4 12.7 98.4 13.6C98.4 14.5 98.5 15.3 98.8 15.9C99.1 16.5 99.5 17 100 17.3C100.5 17.6 101.1 17.8 101.8 17.8C102.4 17.8 103 17.6 103.5 17.3C104 17 104.4 16.5 104.7 15.9L106.9 16.9C106.5 17.9 105.8 18.6 104.9 19.2C104 19.9 103 20.2 101.8 20.2Z" fill="var(--text-dark)"/>
                            <path d="M113.4 20.2C112.2 20.2 111.2 20 110.3 19.5C109.4 19 108.8 18.3 108.3 17.5C107.8 16.6 107.6 15.7 107.6 14.6C107.6 13.5 107.8 12.5 108.3 11.7C108.8 10.8 109.4 10.2 110.3 9.7C111.2 9.2 112.2 9 113.4 9C114.6 9 115.6 9.2 116.5 9.7C117.4 10.2 118 10.8 118.5 11.7C119 12.5 119.2 13.5 119.2 14.6C119.2 15.7 119 16.6 118.5 17.5C118 18.3 117.4 19 116.5 19.5C115.6 20 114.6 20.2 113.4 20.2ZM113.4 18.2C114.1 18.2 114.7 18 115.2 17.7C115.7 17.4 116.1 16.9 116.4 16.3C116.7 15.7 116.8 15 116.8 14.2C116.8 13.4 116.7 12.7 116.4 12.1C116.1 11.5 115.7 11 115.2 10.7C114.7 10.4 114.1 10.2 113.4 10.2C112.7 10.2 112.1 10.4 111.6 10.7C111.1 11 110.7 11.5 110.4 12.1C110.1 12.7 110 13.4 110 14.2C110 15 110.1 15.7 110.4 16.3C110.7 16.9 111.1 17.4 111.6 17.7C112.1 18 112.7 18.2 113.4 18.2Z" fill="var(--text-dark)"/>
                        </svg>
                    </div>
                </Link>

                <div className="header-actions">
                    {currentUser ? (
                        <div className="user-info">
                            <div className="user-avatar">
                                {getInitials(currentUser.username)}
                            </div>
                            <span className="username">{currentUser.username}</span>
                            <button className="header-button" onClick={handleLogout}>Logout</button>
                        </div>
                    ) : (
                        <Link to="/auth" className="header-button">Login</Link>
                    )}
                </div>
            </div>
        </header>
    );
};

export default Header;