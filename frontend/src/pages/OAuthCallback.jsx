import React, { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { handleOAuthCallback } from '../api/authApi';
import { useAuth } from '../contexts/AuthContext';

const OAuthCallback = () => {
    const navigate = useNavigate();
    const { login } = useAuth();

    useEffect(() => {
        const processCallback = async () => {
            try {
                const user = await handleOAuthCallback();
                if (user) {
                    login(user);
                    navigate('/');
                } else {
                    navigate('/auth', { state: { error: 'Authentication failed' } });
                }
            } catch (error) {
                console.error('OAuth callback error:', error);
                navigate('/auth', { state: { error: 'Authentication failed' } });
            }
        };

        processCallback();
    }, [navigate, login]);

    return (
        <div className="loading-container">
            <div className="spinner"></div>
            <p>Completing authentication...</p>
        </div>
    );
};

export default OAuthCallback;