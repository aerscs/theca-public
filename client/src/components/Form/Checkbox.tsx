import { cn } from "@/lib/utils";
import { Check } from "lucide-react";

interface CheckboxProps extends React.InputHTMLAttributes<HTMLInputElement> {
  isChecked?: boolean;
}

export const Checkbox: React.FC<CheckboxProps> = ({
  isChecked,
  children,
  ...props
}) => {
  return (
    <label className="flex cursor-pointer items-center gap-2 py-1">
      <input className="peer" type="checkbox" checked={isChecked} {...props} />
      <span
        aria-checked={isChecked}
        role="checkbox"
        className={cn(
          "border-smoke-500 peer-focus:border-smoke-300 flex h-8 w-8 items-center justify-center gap-2 rounded-lg border-2 p-1 transition-all focus:outline-none",
          isChecked ? "bg-smoke-500/100" : "bg-smoke-500/0",
          props.className,
        )}
      >
        {isChecked && (
          <Check
            width={20}
            height={20}
            strokeWidth={2.5}
            className="text-smoke-100"
          />
        )}
      </span>
      {children}
    </label>
  );
};
