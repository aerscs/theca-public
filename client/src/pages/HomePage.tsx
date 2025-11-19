import { Button } from "@/components/ui/Button";
import { ThecaLogo } from "@/components/Theca";
import { Download, Upload } from "lucide-react";
import { BookmarkFeed } from "@/components/Bookmarks/BookmarkFeed";
import { BookmarkAddModal } from "@/components/Bookmarks/CreateBookmarkModal";
import { BookmarksProvider } from "@/components/Providers/BookmarksProvider";
import { Modal } from "@/components/ui/Modal";
import { SearchBar } from "@/components/ui/SearchBar";
import { Profile } from "@/components/ui/Profile";

export const HomePage = () => {
  return (
    <BookmarksProvider>
      <main className="flex h-full flex-col items-center justify-center">
        <div className="flex flex-col items-center gap-3">
          <SearchBar />

          <div className="relative flex w-full items-center justify-center">
            <p className="bg-smoke-900 text-smoke-300 px-4 py-2 font-medium">
              Bookmarks
            </p>
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="300"
              height="1"
              viewBox="0 0 300 1"
              fill="none"
              className="absolute top-1/2 left-1/2 -z-10 w-full -translate-x-1/2 -translate-y-1/2"
            >
              <path d="M0 1H300" className="stroke-smoke-300" />
            </svg>
          </div>

          <BookmarkFeed />
        </div>
      </main>
      <nav className="border-smoke-700 bg-smoke-900 relative my-2 flex w-full items-center justify-between rounded-xl border-2 p-0.5 sm:my-10">
        <Button text="Theca" className="pl-2">
          <ThecaLogo width={20} height={20} />
        </Button>

        <div className="absolute top-1/2 left-1/2 flex -translate-x-1/2 -translate-y-1/2 gap-0.5">
          <Button icon={Upload} />
          <Modal trigger={<Button text={"Add a bookmark"} />}>
            <BookmarkAddModal className="absolute bottom-12 left-1/2 -translate-x-1/2" />
          </Modal>
          <Button icon={Download} />
        </div>

        <Profile />
      </nav>
    </BookmarksProvider>
  );
};
