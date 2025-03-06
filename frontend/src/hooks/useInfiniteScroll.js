import { useInfiniteQuery } from "react-query";
import { fetchJokes } from "../api/jokesApi";

export const useInfiniteScroll = (sortParams = { sortField: "created_at", order: "desc" }) => {
  return useInfiniteQuery(
      ["jokes", sortParams],
      ({ pageParam }) => fetchJokes({
        pageParam,
        sortField: sortParams.sortField,
        order: sortParams.order
      }),
      {
        getNextPageParam: (lastPage, pages) => pages.length + 1,
      }
  );
};