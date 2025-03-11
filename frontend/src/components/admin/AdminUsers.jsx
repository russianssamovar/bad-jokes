import React, { useState, useEffect } from 'react';
import { getUsers, setUserAdminStatus } from '../../api/adminApi';
import { useAuth } from '../../contexts/AuthContext';
import './AdminStyles.css';
import {Navigate} from "react-router-dom";

const AdminUsers = () => {
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [page, setPage] = useState(1);
    const [totalPages, setTotalPages] = useState(1);
    const [pageSize, setPageSize] = useState(20);

    const auth = useAuth() || {};
    const currentUser = auth.user || null;

    if (!currentUser || !currentUser.isAdmin) {
        return <Navigate to="/" replace />;
    }

    
    const fetchUsers = async () => {
        try {
            setLoading(true);
            const data = await getUsers(page, pageSize);
            setUsers(data.users);
            setTotalPages(data.total_pages);
            setError(null);
        } catch (err) {
            setError('Failed to fetch users');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchUsers();
    }, [page, pageSize]);

    const handleToggleAdmin = async (userId, currentStatus) => {
        try {
            await setUserAdminStatus(userId, !currentStatus);
            setUsers(users.map(user =>
                user.id === userId
                    ? { ...user, is_admin: !currentStatus }
                    : user
            ));
        } catch (err) {
            console.error('Failed to update admin status:', err);
            alert('Failed to update admin status');
        }
    };

    if (loading && users.length === 0) {
        return <div className="admin-loading">Loading users...</div>;
    }

    if (error) {
        return <div className="admin-error">{error}</div>;
    }

    return (
        <div className="admin-users">
            <h2>Users Management</h2>

            <table className="admin-table">
                <thead>
                <tr>
                    <th>ID</th>
                    <th>Username</th>
                    <th>Email</th>
                    <th>Status</th>
                    <th>Created</th>
                    <th>Actions</th>
                </tr>
                </thead>
                <tbody>
                {users.map(user => (
                    <tr key={user.id}>
                        <td>{user.id}</td>
                        <td>{user.username}</td>
                        <td>{user.email}</td>
                        <td>
                            <span className={user.is_admin ? 'status-admin' : 'status-user'}>
                                {user.is_admin ? 'Admin' : 'User'}
                            </span>
                        </td>
                        <td>{new Date(user.created_at).toLocaleString()}</td>
                        <td>
                            {currentUser && user.id !== currentUser.userId && (
                                <button
                                    className="admin-button"
                                    onClick={() => handleToggleAdmin(user.id, user.is_admin)}
                                >
                                    {user.is_admin ? 'Remove Admin' : 'Make Admin'}
                                </button>
                            )}
                        </td>
                    </tr>
                ))}
                </tbody>
            </table>

            <div className="pagination">
                <button
                    disabled={page === 1}
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                >
                    Previous
                </button>
                <span>Page {page} of {totalPages}</span>
                <button
                    disabled={page === totalPages}
                    onClick={() => setPage(p => Math.min(totalPages, p + 1))}
                >
                    Next
                </button>
            </div>
        </div>
    );
};

export default AdminUsers;