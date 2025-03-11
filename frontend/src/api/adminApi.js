import { api } from '../utils/api';

export const getUsers = async (page = 1, pageSize = 20) => {
    const response = await api.get(`/admin/users?page=${page}&page_size=${pageSize}`);
    return response.data;
};

export const setUserAdminStatus = async (userId, isAdmin) => {
    const response = await api.post('/admin/users/admin-status', { user_id: userId, is_admin: isAdmin });
    return response.data;
};

export const getModerationLogs = async (page = 1, pageSize = 50) => {
    const response = await api.get(`/admin/logs?page=${page}&page_size=${pageSize}`);
    return response.data;
};

export const getUserStats = async () => {
    const response = await api.get('/admin/stats');
    return response.data;
};

export const deleteAsAdminJoke = async (jokeId) => {
    await api.delete(`/admin/jokes/${jokeId}`);
};

export const deleteAsAdminComment = async (commentId) => {
    await api.delete(`/admin/comments/${commentId}`);
};