import { createContext, useContext } from "react";

export interface BookmarkType {
  favicon?: string;
  id?: number;
  show_text: boolean;
  title: string;
  url: string;
}

interface BookmarksContextType {
  bookmarks: BookmarkType[];
  createBookmark: (bookmark: BookmarkType) => Promise<void>;
  readBookmarks: () => Promise<void>;
  updateBookmark: (bookmark: BookmarkType) => Promise<void>;
  deleteBookmark: (id: number) => void;
  isLoading: boolean;
  error: string | null;
}

export const BookmarksContext = createContext<BookmarksContextType | undefined>(
  undefined,
);

export const useBookmarks = () => {
  const context = useContext(BookmarksContext);
  if (!context) {
    throw new Error("useBookmarks must be used within an BookmarksProvider");
  }
  return context;
};
