import React from "react";
import JokeList from "../components/JokeList";

const Home = () => (
  <div className="min-h-screen bg-gray-100 p-4">
    <h1 className="text-2xl font-bold text-center mb-4">Jokes</h1>
    <JokeList />
  </div>
);

export default Home;
