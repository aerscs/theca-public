import React from "react";
import { cn } from "@/lib/utils";
import type { LucideIcon } from "lucide-react";

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  size?: "sm" | "md";
  icon?: LucideIcon;
  text?: string;
  reversed?: boolean;
  children?: React.ReactNode;
  className?: string;
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      icon: Icon,
      text,
      children,
      className,
      reversed = false,
      size = "sm",
      ...props
    },
    ref,
  ) => {
    const iconSize = size === "sm" ? 20 : 24;

    return (
      <button
        className={cn(
          "bg-smoke-700 text-smoke-100 hover:bg-smoke-500 focus-visible:bg-smoke-500 disabled:text-smoke-300 data-[danger]:hover:bg-danger/30 data-[danger]:focus-visible:bg-danger/30 data-[danger]:bg-danger/20 data-[danger]:text-danger flex items-center justify-center rounded-lg font-medium transition-all focus-visible:outline-none disabled:cursor-not-allowed",
          size == "sm"
            ? `${text ? "px-4" : "px-2"} gap-2 py-2 text-[0.875rem] leading-5!`
            : `${text ? "px-4.5" : "px-2.5"} py-2.5 text-[1.125rem]`,
          reversed ? "flex-row-reverse" : "flex-row",
          className,
        )}
        ref={ref}
        {...props}
      >
        {Icon ? <Icon size={iconSize} /> : children}
        {text && <p className="truncate">{text}</p>}
      </button>
    );
  },
);
