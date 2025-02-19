import React from "react";

const JokeBody = ({ joke }) => (
  <div>
    <h2>{joke.title}</h2>
    <p>{joke.body}</p>
  </div>
);

export default JokeBody; 