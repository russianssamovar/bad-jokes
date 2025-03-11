import React, { useRef, useEffect } from "react";
import JokeCard from "./JokeCard";
import { useInfiniteScroll } from "../hooks/useInfiniteScroll";

const JokeList = ({ sortParams }) => {
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, refetch } = useInfiniteScroll(sortParams);
  const bottomRef = useRef(null);

  useEffect(() => {
    if (!bottomRef.current) return;

    const observer = new IntersectionObserver(([entry]) => {
      if (entry.isIntersecting && hasNextPage && data?.pages?.[data.pages.length - 1] !== null) {
        fetchNextPage();
      }
    }, { threshold: 1.0 });

    observer.observe(bottomRef.current);
    return () => observer.disconnect();
  }, [fetchNextPage, hasNextPage, data]);

  const handleJokeDelete = (deletedJokeId) => {
    refetch();
  };

  return (
      <div className="joke-list-container">
        {data?.pages
            ? data.pages.map((page) =>
                page ? page.map((joke) => (
                    <JokeCard
                        key={joke.id}
                        joke={joke}
                        onDelete={handleJokeDelete}
                    />
                )) : null
            )
            : null}
        <div ref={bottomRef} className="load-more-trigger"></div>
        {isFetchingNextPage && <p className="loading">Loading more...</p>}
      </div>
  );
};

export default JokeList;