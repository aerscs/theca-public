import { Modal } from "../ui/Modal";
import { BookmarkEditModal } from "./UpdateBookmarkModal";

interface BookmarkProps extends React.HTMLAttributes<HTMLAnchorElement> {
  bookmardId: number;
  title: string;
  url: string;
  show_text: boolean;
  favicon?: string;
}

export const Bookmark: React.FC<BookmarkProps> = ({
  bookmardId,
  title,
  url,
  show_text,
  favicon,
  ...props
}) => {
  const fallbackFavicon =
    "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNCIgaGVpZ2h0PSIyNCIgdmlld0JveD0iMCAwIDI0IDI0IiBmaWxsPSJub25lIj4KICA8cGF0aCBkPSJNMC41IDUuNDI4NTdDMC41IDIuNzA2NiAyLjcwNjYgMC41IDUuNDI4NTcgMC41SDE4LjU3MTRDMjEuMjkzNCAwLjUgMjMuNSAyLjcwNjYgMjMuNSA1LjQyODU3VjE4LjU3MTRDMjMuNSAyMS4yOTM0IDIxLjI5MzQgMjMuNSAxOC41NzE0IDIzLjVINS40Mjg1N0MyLjcwNjYgMjMuNSAwLjUgMjEuMjkzNCAwLjUgMTguNTcxNFY1LjQyODU3WiIgZmlsbD0iIzI0MjQzMyIvPgogIDxwYXRoIGQ9Ik0xOCAxOC4xMDRDMTggMTguNzMwMyAxNy4zMjcyIDE5LjEyNjIgMTYuNzc5NyAxOC44MjJMMTIuMzk4OSAxNi4zODgzQzEyLjE1MDggMTYuMjUwNSAxMS44NDkyIDE2LjI1MDUgMTEuNjAxMSAxNi4zODgzTDcuMjIwMzUgMTguODIyQzYuNjcyODQgMTkuMTI2MiA2IDE4LjczMDMgNiAxOC4xMDRWNi4xNjY2N0M2IDUuNzI0NjQgNi4xODA2MSA1LjMwMDcyIDYuNTAyMSA0Ljk4ODE2QzYuODIzNTkgNC42NzU1OSA3LjI1OTYzIDQuNSA3LjcxNDI5IDQuNUgxNi4yODU3QzE2Ljc0MDQgNC41IDE3LjE3NjQgNC42NzU1OSAxNy40OTc5IDQuOTg4MTZDMTcuODE5NCA1LjMwMDcyIDE4IDUuNzI0NjQgMTggNi4xNjY2N1YxOC4xMDRaIiBmaWxsPSIjNDE0MTU4Ii8+Cjwvc3ZnPg==";

  return (
    <Modal
      rightClick
      trigger={
        <a
          href={url}
          className={`border-smoke-700 group flex w-fit items-center justify-center gap-2 rounded-lg border-2 px-2 ${show_text && "pr-4"} hover:bg-smoke-700 bg-smoke-900 focus:bg-smoke-700 py-2 transition-all focus:outline-none`}
          target="_blank"
          rel="noreferrer"
          title={title}
          {...props}
        >
          <img
            src={favicon || fallbackFavicon}
            alt={`${title} favicon`}
            className="h-6 w-6 opacity-50 transition-all group-hover:opacity-100 group-focus:opacity-100"
          />
          {show_text ? (
            <p className="group-hover:text-smoke-100 group-focus:text-smoke-100 text-smoke-300 text-[18px] transition-all">
              {title}
            </p>
          ) : (
            ""
          )}
        </a>
      }
    >
      <BookmarkEditModal
        bookmarkId={bookmardId}
        title={title}
        url={url}
        show_text={show_text}
        className="absolute top-12 left-1/2 -translate-x-1/2"
      />
    </Modal>
  );
};
