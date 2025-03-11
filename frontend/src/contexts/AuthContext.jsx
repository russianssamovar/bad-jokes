import React, { createContext, useContext, useState, useEffect } from 'react';
import { getCurrentUser, loginUser, logoutUser } from '../api/authApi';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
    const [user, setUser] = useState(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const currentUser = getCurrentUser();
        if (currentUser) {
            setUser(currentUser);
        }
        setIsLoading(false);
    }, []);

    const login = async (email, password) => {
        try {
            const response = await loginUser(email, password);
            localStorage.setItem('token', response.token);
            const user = getCurrentUser();
            setUser(user);
            return user;
        } catch (error) {
            throw error;
        }
    };

    const logout = () => {
        logoutUser();
        setUser(null);
    };

    const value = {
        user,
        login,
        logout,
        isLoading
    };

    return (
        <AuthContext.Provider value={value}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    return useContext(AuthContext);
}