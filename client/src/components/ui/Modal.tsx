import { useOutsideClick } from "@/hooks/useOutsideClick";
import { cn } from "@/lib/utils";
import { useRef, useState } from "react";

interface ModalProps extends React.HTMLAttributes<HTMLDivElement> {
  trigger: React.ReactNode;
  rightClick?: boolean;
}

export const Modal: React.FC<ModalProps> = ({
  rightClick = false,
  trigger,
  children,
  className,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const modalRef = useRef<HTMLDivElement>(null);

  useOutsideClick(modalRef, () => setIsOpen(false));

  const handler = rightClick
    ? {
        onContextMenu: (e: React.MouseEvent) => {
          e.preventDefault();
          setIsOpen(true);
        },
      }
    : { onClick: () => setIsOpen(true) };

  return (
    <div
      className={cn("relative inline-block", className)}
      aria-expanded={isOpen}
      ref={modalRef}
    >
      <span {...handler}>{trigger}</span>

      {isOpen && children}
    </div>
  );
};
