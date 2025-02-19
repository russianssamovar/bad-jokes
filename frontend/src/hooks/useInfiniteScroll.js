import { useInfiniteQuery } from "react-query";
import { fetchJokes } from "../api/jokesApi";

export const useInfiniteScroll = () => {
  return useInfiniteQuery("jokes", fetchJokes, {
    getNextPageParam: (lastPage, pages) => pages.length + 1,
  });
};
