import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { loginUser, registerUser } from "../api/authApi";
import OAuthButtons from "../components/OAuthButtons.jsx";

const AuthPage = () => {
  const [activeTab, setActiveTab] = useState("login");
  const [formData, setFormData] = useState({
    login: { email: "", password: "" },
    register: { email: "", password: "", username: "" }
  });
  const [validationErrors, setValidationErrors] = useState({
    username: [],
    password: []
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

    if (tab === 'register' && (field === 'username' || field === 'password')) {
      validateField(field, value);
    }
  };

  const validateField = (field, value) => {
    const errors = [];

    if (field === 'username') {
      if (value.length < 3) {
        errors.push("Username must be at least 3 characters long");
      }
      if (value.length > 20) {
        errors.push("Username must be at most 20 characters long");
      }
      if (value.includes(" ")) {
        errors.push("Username cannot contain spaces");
      }
      if (!/^[a-zA-Z0-9._-]+$/.test(value)) {
        errors.push("Username can only contain letters, numbers, dots, underscores, and hyphens");
      }

      const forbiddenWords = ["admin", "administrator", "mod", "moderator", "system", "support",
        "staff", "official", "root", "superuser"];

      const lowerValue = value.toLowerCase();
      for (const word of forbiddenWords) {
        if (lowerValue.includes(word)) {
          errors.push(`Username cannot contain '${word}'`);
          break;
        }
      }
    }

    if (field === 'password') {
      if (value.length < 8) {
        errors.push("Password must be at least 8 characters long");
      }
      if (value.length > 72) {
        errors.push("Password must be at most 72 characters long");
      }
      if (!/[A-Z]/.test(value)) {
        errors.push("Password must contain an uppercase letter");
      }
      if (!/[a-z]/.test(value)) {
        errors.push("Password must contain a lowercase letter");
      }
      if (!/[0-9]/.test(value)) {
        errors.push("Password must contain a number");
      }
      if (!/[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?~]/.test(value)) {
        errors.push("Password must contain a special character");
      }

      const commonPasswords = ["password", "123456", "qwerty", "12345678", "111111", "1234567890", "password123", "admin", "welcome", "abc123"];
      if (commonPasswords.includes(value.toLowerCase())) {
        errors.push("This password is too common");
      }
    }

    setValidationErrors(prev => ({
      ...prev,
      [field]: errors
    }));

    return errors.length === 0;
  };

  const validateForm = () => {
    const usernameValid = validateField('username', formData.register.username);
    const passwordValid = validateField('password', formData.register.password);

    return usernameValid && passwordValid;
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

    if (!validateForm()) {
      setError("Please correct the form errors before submitting");
      return;
    }

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
      // Try to parse error message from response if available
      if (err.response && err.response.data && err.response.data.error) {
        setError(err.response.data.error);
      } else {
        setError("Registration failed. Please try again.");
      }
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    setValidationErrors({
      username: [],
      password: []
    });
  }, [activeTab]);

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
          
          <OAuthButtons />
          
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
                {formData.register.username && validationErrors.username.length > 0 && (
                    <div className="validation-errors">
                      {validationErrors.username.map((err, index) => (
                          <div key={index} className="validation-error">{err}</div>
                      ))}
                    </div>
                )}

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
                {formData.register.password && validationErrors.password.length > 0 && (
                    <div className="validation-errors">
                      {validationErrors.password.map((err, index) => (
                          <div key={index} className="validation-error">{err}</div>
                      ))}
                    </div>
                )}

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