import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { loginUser, registerUser } from "../api/authApi";

const AuthPage = () => {
  const [activeTab, setActiveTab] = useState("login");
  const [formData, setFormData] = useState({
    login: { email: "", password: "" },
    register: { email: "", password: "", username: "" }
  });
  const [error, setError] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();

  const handleInputChange = (tab, field, value) => {
    setFormData(prev => ({
      ...prev,
      [tab]: { ...prev[tab], [field]: value }
    }));
    setError("");
  };

  const handleLogin = async (e) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const { token } = await loginUser(formData.login.email, formData.login.password);
      localStorage.setItem("token", token);
      navigate("/");
    } catch (err) {
      setError("Invalid email or password");
    } finally {
      setIsLoading(false);
    }
  };

  const handleRegister = async (e) => {
    e.preventDefault();
    setIsLoading(true);

    try {
      const { token } = await registerUser(
          formData.register.username,
          formData.register.email,
          formData.register.password
      );
      localStorage.setItem("token", token);
      navigate("/");
    } catch (err) {
      setError("Registration failed. Please try again.");
    } finally {
      setIsLoading(false);
    }
  };

  return (
      <div className="auth-container">
        <div className="auth-card">
          <div className="auth-header">
            <h1>Welcome Back</h1>
            <p>Join the community of joke lovers</p>
          </div>

          <div className="auth-tabs">
            <button
                className={`auth-tab ${activeTab === 'login' ? 'active' : ''}`}
                onClick={() => setActiveTab('login')}
            >
              Login
            </button>
            <button
                className={`auth-tab ${activeTab === 'register' ? 'active' : ''}`}
                onClick={() => setActiveTab('register')}
            >
              Register
            </button>
            <div className="tab-indicator" style={{ left: activeTab === 'login' ? '0%' : '50%' }}></div>
          </div>

          <div className="forms-container">
            <div className={`form-section ${activeTab === 'login' ? 'active' : ''}`}>
              <form onSubmit={handleLogin}>
                <div className="input-group">
                  <input
                      type="email"
                      value={formData.login.email}
                      onChange={(e) => handleInputChange('login', 'email', e.target.value)}
                      required
                  />
                  <label>Email</label>
                  <i className="input-icon">✉️</i>
                </div>

                <div className="input-group">
                  <input
                      type="password"
                      value={formData.login.password}
                      onChange={(e) => handleInputChange('login', 'password', e.target.value)}
                      required
                  />
                  <label>Password</label>
                  <i className="input-icon">🔒</i>
                </div>

                <button type="submit" className="auth-button" disabled={isLoading}>
                  {isLoading ? 'Logging in...' : 'Login'}
                </button>
              </form>
            </div>

            <div className={`form-section ${activeTab === 'register' ? 'active' : ''}`}>
              <form onSubmit={handleRegister}>
                <div className="input-group">
                  <input
                      type="text"
                      value={formData.register.username}
                      onChange={(e) => handleInputChange('register', 'username', e.target.value)}
                      required
                  />
                  <label>Username</label>
                  <i className="input-icon">👤</i>
                </div>

                <div className="input-group">
                  <input
                      type="email"
                      value={formData.register.email}
                      onChange={(e) => handleInputChange('register', 'email', e.target.value)}
                      required
                  />
                  <label>Email</label>
                  <i className="input-icon">✉️</i>
                </div>

                <div className="input-group">
                  <input
                      type="password"
                      value={formData.register.password}
                      onChange={(e) => handleInputChange('register', 'password', e.target.value)}
                      required
                  />
                  <label>Password</label>
                  <i className="input-icon">🔒</i>
                </div>

                <button type="submit" className="auth-button" disabled={isLoading}>
                  {isLoading ? 'Creating account...' : 'Create Account'}
                </button>
              </form>
            </div>
          </div>

          {error && <div className="error-message">{error}</div>}
        </div>
      </div>
  );
};

export default AuthPage;