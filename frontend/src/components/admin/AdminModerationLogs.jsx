import React, { useState, useEffect } from 'react';
import { getModerationLogs } from '../../api/adminApi';
import './AdminStyles.css';
import { useAuth } from "../../contexts/AuthContext.jsx";
import { Navigate } from "react-router-dom";

const AdminModerationLogs = () => {
    const [logs, setLogs] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [page, setPage] = useState(1);
    const [pageSize, setPageSize] = useState(50);
    const auth = useAuth() || {};
    const currentUser = auth.user || null;

    const fetchLogs = async () => {
        try {
            setLoading(true);
            const data = await getModerationLogs(page, pageSize);
            setLogs(data || []); // Ensure logs is never null
            setError(null);
        } catch (err) {
            setError('Failed to fetch moderation logs');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (currentUser?.isAdmin) {
            fetchLogs();
        }
    }, [page, pageSize, currentUser?.isAdmin]);

    if (!currentUser || !currentUser.isAdmin) {
        return <Navigate to="/" replace />;
    }

    if (loading && !logs?.length) {
        return <div className="admin-loading">Loading moderation logs...</div>;
    }

    if (error) {
        return <div className="admin-error">{error}</div>;
    }

    return (
        <div className="admin-mod-logs">
            <h2>Moderation Logs</h2>

            {loading ? (
                <div className="admin-loading">Loading logs...</div>
            ) : error ? (
                <div className="admin-error">Error loading logs: {error}</div>
            ) : logs && logs.length > 0 ? (
                <table className="admin-table">
                    <thead>
                    <tr>
                        <th>ID</th>
                        <th>Admin</th>
                        <th>Action</th>
                        <th>Target ID</th>
                        <th>Date</th>
                    </tr>
                    </thead>
                    <tbody>
                    {logs.map(log => (
                        <tr key={log.id}>
                            <td>{log.id}</td>
                            <td>{log.admin_username}</td>
                            <td><span className={`action-${log.action}`}>{log.action}</span></td>
                            <td>{log.target_id}</td>
                            <td>{new Date(log.created_at).toLocaleString()}</td>
                        </tr>
                    ))}
                    </tbody>
                </table>
            ) : (
                <div className="admin-notice">No logs found.</div>
            )}

            <div className="pagination">
                <button
                    disabled={page === 1}
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                >
                    Previous
                </button>
                <span>Page {page}</span>
                <button onClick={() => setPage(p => p + 1)}>
                    Next
                </button>
            </div>
        </div>
    );
};

export default AdminModerationLogs;