import React from "react";
import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "react-query";
import Home from "./pages/Home";
import Login from "./pages/Login";
import Header from "./components/Header";
import CreateJoke from "./components/CreateJoke";
import JokeDetail from './pages/JokeDetail';

const queryClient = new QueryClient();

const App = () => (
  <QueryClientProvider client={queryClient}>
    <Router>
      <Header />
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/login" element={<Login />} />
        <Route path="/create" element={<CreateJoke />} />
        <Route path="/joke/:jokeId" element={<JokeDetail />} />
      </Routes>
    </Router>
  </QueryClientProvider>
);

export default App;
