import React, { useState, useEffect } from 'react';
import { getUserStats } from '../../api/adminApi';
import './AdminStyles.css';
import {useAuth} from "../../contexts/AuthContext.jsx";
import {Navigate} from "react-router-dom";

const AdminStats = () => {
    const [stats, setStats] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const auth = useAuth() || {};
    const currentUser = auth.user || null;

    if (!currentUser || !currentUser.isAdmin) {
        return <Navigate to="/" replace />;
    }

    const fetchStats = async () => {
        try {
            setLoading(true);
            const data = await getUserStats();
            setStats(data);
            setError(null);
        } catch (err) {
            setError('Failed to fetch statistics');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchStats();
    }, []);

    if (loading) {
        return <div className="admin-loading">Loading statistics...</div>;
    }

    if (error) {
        return <div className="admin-error">{error}</div>;
    }

    if (!stats) {
        return <div className="admin-error">No statistics available</div>;
    }

    return (
        <div className="admin-stats">
            <h2>User Statistics</h2>

            <div className="stats-grid">
                <div className="stat-card">
                    <h3>Total Users</h3>
                    <div className="stat-value">{stats.total_users}</div>
                </div>

                <div className="stat-card">
                    <h3>Admin Users</h3>
                    <div className="stat-value">{stats.admin_count}</div>
                </div>

                <div className="stat-card">
                    <h3>New Users (24h)</h3>
                    <div className="stat-value">{stats.new_users_today}</div>
                </div>

                <div className="stat-card">
                    <h3>New Users (Week)</h3>
                    <div className="stat-value">{stats.new_users_this_week}</div>
                </div>

                <div className="stat-card">
                    <h3>New Users (Month)</h3>
                    <div className="stat-value">{stats.new_users_this_month}</div>
                </div>
            </div>

            <h3>Most Active Users</h3>
            <table className="admin-table">
                <thead>
                <tr>
                    <th>User ID</th>
                    <th>Username</th>
                    <th>Jokes</th>
                    <th>Comments</th>
                    <th>Total Activity</th>
                </tr>
                </thead>
                <tbody>
                {stats.most_active_users.map(user => (
                    <tr key={user.id}>
                        <td>{user.id}</td>
                        <td>{user.username}</td>
                        <td>{user.jokes_count}</td>
                        <td>{user.comments_count}</td>
                        <td>{user.jokes_count + user.comments_count}</td>
                    </tr>
                ))}
                </tbody>
            </table>
        </div>
    );
};

export default AdminStats;