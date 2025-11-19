import { cn } from "@/lib/utils";
import { forwardRef } from "react";

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  first?: boolean;
  inputError?: string;
  className?: string;
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ inputError, first, className, ...props }, ref) => {
    return (
      <div
        className={cn(
          "bg-smoke-900 relative w-full transition-all",
          !first ? (!inputError ? "mt-0" : "not-focus-within:mt-1") : "",
        )}
      >
        <input
          ref={ref}
          className={cn(
            "peer text-smoke-100 placeholder:text-smoke-300 focus:border-smoke-300 w-full rounded-lg border-2 bg-transparent px-4 py-2.5 transition-all outline-none focus:outline-none",
            inputError ? "border-danger" : "border-smoke-500",
            className,
          )}
          {...props}
        />
        {inputError && (
          <span
            role="alert"
            aria-live="assertive"
            className={cn(
              "bg-smoke-900 text-danger absolute -top-3 left-3 flex px-1 py-0.5 text-[0.75rem] transition-all",
              "peer-focus:opacity-0",
              "opacity-100",
            )}
          >
            {inputError}
          </span>
        )}
      </div>
    );
  },
);
