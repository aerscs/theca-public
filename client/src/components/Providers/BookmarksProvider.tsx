import { useCallback, useEffect, useState } from "react";
import { BookmarksContext, type BookmarkType } from "@/hooks/useBookmarks";
import { api } from "@/api/axiosInstance";

export const BookmarksProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [bookmarks, setBookmarks] = useState<BookmarkType[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const createBookmark = async (bookmark: BookmarkType) => {
    try {
      const response = await api.post("/v1/api/bookmarks", bookmark);
      setBookmarks((prevBookmarks) => [...prevBookmarks, response.data.data]);
    } catch (error) {
      console.error("Error creating bookmark:", error);
      setError("Failed to create bookmark.");
    } finally {
      setIsLoading(false);
    }
  };

  const readBookmarks = useCallback(async () => {
    try {
      const response = await api.get("/v1/api/bookmarks");
      setBookmarks(response.data.data);
    } catch (error) {
      console.error("Error fetching bookmarks:", error);
      setError("Failed to load items.");
    } finally {
      setIsLoading(false);
    }
  }, []);

  const updateBookmark = async (bookmark: BookmarkType) => {
    try {
      const response = await api.patch(
        `/v1/api/bookmarks/${bookmark.id}`,
        bookmark,
      );
      setBookmarks((prevBookmarks) =>
        prevBookmarks.map((item) =>
          item.id === bookmark.id ? response.data.data : item,
        ),
      );
    } catch (error) {
      console.error("Error updating bookmark:", error);
      setError("Failed to update bookmark.");
    } finally {
      setIsLoading(false);
    }
  };

  const deleteBookmark = async (bookmarkId: number) => {
    try {
      await api.delete(`/v1/api/bookmarks/${bookmarkId}`);
      setBookmarks((prevBookmarks) =>
        prevBookmarks.filter((item) => item.id !== bookmarkId),
      );
    } catch (error) {
      console.error("Error deleting bookmark:", error);
      setError("Failed to delete bookmark.");
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    readBookmarks();
  }, [readBookmarks]);

  return (
    <BookmarksContext.Provider
      value={{
        bookmarks,
        createBookmark,
        readBookmarks,
        updateBookmark,
        deleteBookmark,
        isLoading,
        error,
      }}
    >
      {children}
    </BookmarksContext.Provider>
  );
};
