import { cn } from "../../lib/utils";

interface FormProps extends React.FormHTMLAttributes<HTMLFormElement> {
  formError?: string;
  className?: string;
}

export const Form: React.FC<FormProps> = ({
  formError,
  className,
  children,
  ...props
}) => {
  return (
    <form
      className={cn(
        "border-smoke-700 bg-smoke-900 group data-[danger=true]:border-danger relative flex w-[360px] flex-col items-start justify-center gap-2 rounded-[1.25rem] border-2 p-3 transition-all",
        className,
      )}
      data-danger={formError ? true : false}
      {...props}
    >
      {children}
      <label
        className={cn(
          "bg-smoke-900 text-danger absolute -top-3 left-3 flex px-1 py-0.5 text-[0.75rem] opacity-0 transition-all group-data-[danger=true]:opacity-100 peer-focus:opacity-0",
        )}
      >
        {formError}
      </label>
    </form>
  );
};
