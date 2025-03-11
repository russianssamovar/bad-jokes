import React, { useState } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { Navigate } from 'react-router-dom';
import AdminUsers from './AdminUsers';
import AdminStats from './AdminStats';
import AdminModerationLogs from './AdminModerationLogs';
import './AdminStyles.css';

const AdminPanel = () => {
    const [activeView, setActiveView] = useState('dashboard');
    const auth = useAuth() || {};
    const { user: currentUser, isLoading } = auth;

    if (isLoading) {
        return <div className="loading-spinner">Loading admin panel...</div>;
    }
    
    if (!currentUser || !currentUser.isAdmin) {
        return <Navigate to="/" replace />;
    }

    const renderContent = () => {
        switch (activeView) {
            case 'users':
                return <AdminUsers />;
            case 'stats':
                return <AdminStats />;
            case 'logs':
                return <AdminModerationLogs />;
            default:
                return (
                    <div className="admin-welcome">
                        <h2>Welcome to Admin Dashboard</h2>
                        <p>Select an option from the sidebar to manage your application.</p>

                        <div className="admin-quick-links">
                            <div className="admin-card">
                                <h3>Users Management</h3>
                                <p>View and manage user accounts, change admin status</p>
                                <button onClick={() => setActiveView('users')} className="admin-button">Manage Users</button>
                            </div>

                            <div className="admin-card">
                                <h3>Statistics</h3>
                                <p>View user activity statistics and analytics</p>
                                <button onClick={() => setActiveView('stats')} className="admin-button">View Stats</button>
                            </div>

                            <div className="admin-card">
                                <h3>Moderation Logs</h3>
                                <p>Review moderation actions and content changes</p>
                                <button onClick={() => setActiveView('logs')} className="admin-button">View Logs</button>
                            </div>
                        </div>
                    </div>
                );
        }
    };

    return (
        <div className="admin-panel">
            <div className="admin-sidebar">
                <h2>Admin Panel</h2>
                <nav>
                    <ul>
                        <li>
                            <button
                                className={activeView === 'dashboard' ? 'active' : ''}
                                onClick={() => setActiveView('dashboard')}
                            >
                                Dashboard
                            </button>
                        </li>
                        <li>
                            <button
                                className={activeView === 'users' ? 'active' : ''}
                                onClick={() => setActiveView('users')}
                            >
                                Users Management
                            </button>
                        </li>
                        <li>
                            <button
                                className={activeView === 'stats' ? 'active' : ''}
                                onClick={() => setActiveView('stats')}
                            >
                                Statistics
                            </button>
                        </li>
                        <li>
                            <button
                                className={activeView === 'logs' ? 'active' : ''}
                                onClick={() => setActiveView('logs')}
                            >
                                Moderation Logs
                            </button>
                        </li>
                    </ul>
                </nav>
            </div>

            <div className="admin-content">
                {renderContent()}
            </div>
        </div>
    );
};

export default AdminPanel;