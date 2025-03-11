import React from "react";
import { AuthProvider } from './contexts/AuthContext';
import "./assets/styles.css";
import { createRoot } from "react-dom/client";
import App from "./App";

const root = document.getElementById("root");
createRoot(root).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
