import { Plus } from "lucide-react";
import { useBookmarks, type BookmarkType } from "@/hooks/useBookmarks";
import { Bookmark } from "./Bookmark";
import { Modal } from "../ui/Modal";
import { Button } from "../ui/Button";
import { BookmarkAddModal } from "./CreateBookmarkModal";

export const BookmarkFeed: React.FC = () => {
  const { bookmarks } = useBookmarks();

  return (
    <div className="flex w-full flex-wrap items-center justify-center gap-1">
      {bookmarks.map((bookmark: BookmarkType) => (
        <Bookmark
          key={bookmark.id}
          title={bookmark.title}
          url={bookmark.url}
          show_text={bookmark.show_text}
          favicon={bookmark.favicon}
          bookmardId={bookmark.id!}
        />
      ))}

      <Modal trigger={<Button icon={Plus} size="md" />}>
        <BookmarkAddModal className="absolute top-12 left-1/2 -translate-x-1/2" />
      </Modal>
    </div>
  );
};
