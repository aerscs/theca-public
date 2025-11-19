import type { SlotProps } from "input-otp";
import { cn } from "../../lib/utils";

export const Slot = (props: SlotProps) => {
  return (
    <div
      className={cn(
        "peer border-smoke-500 text-smoke-100 placeholder:text-smoke-300 focus:border-smoke-300 relative flex h-10 w-full items-center justify-center rounded-lg border-2 bg-transparent px-4 py-2.5 font-medium transition-all outline-none focus:outline-none",
        { "border-smoke-100": props.isActive },
      )}
    >
      <div className="group-has-[input[data-input-otp-placeholder-shown]]:opacity-20">
        {props.char ?? props.placeholderChar}
      </div>
    </div>
  );
};
