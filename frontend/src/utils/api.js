import axios from 'axios';

const BASE_API_URL = import.meta.env.VITE_API_URL || `${window.location.protocol}//${window.location.host}/api`;

export const api = axios.create({
    baseURL: BASE_API_URL,
});

api.interceptors.request.use((config) => {
    const token = localStorage.getItem("token");
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});