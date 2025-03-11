import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "react-query";
import { AuthProvider } from "./contexts/AuthContext"; // Add this import
import Home from "./pages/Home";
import AuthPage from "./pages/AuthPage";
import Header from "./components/Header";
import CreateJoke from "./components/CreateJoke";
import JokeDetail from './pages/JokeDetail';
import AdminUsers from './components/admin/AdminUsers';
import AdminModerationLogs from './components/admin/AdminModerationLogs';
import AdminStats from './components/admin/AdminStats';
import AdminPanel from "./components/admin/AdminPanel";

const queryClient = new QueryClient();

const App = () => (
    <QueryClientProvider client={queryClient}>
        <AuthProvider>
            <Router>
                <Header />
                <Routes>
                    <Route path="/" element={<Home />} />
                    <Route path="/auth" element={<AuthPage />} />
                    <Route path="/create" element={<CreateJoke />} />
                    <Route path="/joke/:jokeId" element={<JokeDetail />} />
                    <Route path="/admin" element={<AdminPanel />} />
                    <Route path="/admin/users" element={<AdminUsers />} />
                    <Route path="/admin/logs" element={<AdminModerationLogs />} />
                    <Route path="/admin/stats" element={<AdminStats />} />
                </Routes>
            </Router>
        </AuthProvider>
    </QueryClientProvider>
);

export default App;