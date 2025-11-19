import { useCallback, useEffect } from "react";

export const useOutsideClick = (
  ref: React.RefObject<HTMLElement | null>,
  callback: () => void,
) => {
  const handleClick = useCallback(
    (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        callback();
      }
    },
    [ref, callback],
  );

  useEffect(() => {
    document.addEventListener("mousedown", handleClick);

    return () => {
      document.removeEventListener("mousedown", handleClick);
    };
  }, [handleClick]);
};
