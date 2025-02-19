import React from "react";

const SignupForm = () => {
  return (
    <div className="auth-form">
      <h2>Signup</h2>
      <input type="email" placeholder="Email" />
      <input type="password" placeholder="Password" />
      <button>Signup</button>
    </div>
  );
};

export default SignupForm;
