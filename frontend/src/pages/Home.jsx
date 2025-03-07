import React, { useState } from "react";
import { Link } from "react-router-dom";
import JokeList from "../components/JokeList";

const Home = () => {
    const [sortParams, setSortParams] = useState({
        sortField: "created_at",
        order: "desc"
    });

    const sortOptions = [
        { label: "Newest", field: "created_at", order: "desc" },
        { label: "Oldest", field: "created_at", order: "asc" },
        { label: "Best", field: "score", order: "desc" },
        { label: "Most Reactions", field: "reactions_count", order: "desc" },
        { label: "Popular", field: "comments_count", order: "desc" }
    ];

    const handleSortChange = (field, order) => {
        setSortParams({ sortField: field, order });
    };

    return (
        <div className="home-container">
            <div className="sorting-panel">
                <div className="sort-buttons">
                    {sortOptions.map((option) => (
                        <button
                            key={option.label}
                            className={`sort-button ${
                                sortParams.sortField === option.field &&
                                sortParams.order === option.order ? 'active' : ''
                            }`}
                            onClick={() => handleSortChange(option.field, option.order)}
                        >
                            {option.label}
                        </button>
                    ))}
                </div>
                <Link to="/create" className="create-button">
                    Create Post
                </Link>
            </div>

            <JokeList sortParams={sortParams} />
        </div>
    );
};

export default Home;