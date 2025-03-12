import React from 'react';
import { FcGoogle } from 'react-icons/fc';
import { FaGithub } from 'react-icons/fa';
import { api } from '../utils/api';

const OAuthButtons = () => {
    const handleOAuthLogin = (provider) => {
        const baseUrl = api.defaults.baseURL;
        const redirectUri = `${window.location.origin}/auth/callback`;
        window.location.href = `${baseUrl}/auth/${provider}/login?redirect_uri=${encodeURIComponent(redirectUri)}`;
    };

    return (
        <div className="oauth-buttons-container">
            <button
                className="oauth-button google-button"
                onClick={() => handleOAuthLogin('google')}
            >
                <FcGoogle size={20} />
                Continue with Google
            </button>

            {/*<button*/}
            {/*    className="oauth-button github-button"*/}
            {/*    onClick={() => handleOAuthLogin('github')}*/}
            {/*>*/}
            {/*    <FaGithub size={20} />*/}
            {/*    Continue with GitHub*/}
            {/*</button>*/}

            <div className="oauth-divider">
                <div className="oauth-divider-line"></div>
                <div className="oauth-divider-text">
                    <span>Or continue with</span>
                </div>
            </div>
        </div>
    );
};

export default OAuthButtons;