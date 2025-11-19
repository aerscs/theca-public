import { Input } from "@/components/Form/Input";
import { Checkbox } from "../Form/Checkbox";
import { useState } from "react";
import { z } from "zod";
import { useForm, type SubmitHandler } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/Button";
import { cn } from "@/lib/utils";
import { useBookmarks } from "@/hooks/useBookmarks";
import { Loader } from "../ui/Loader";

const schema = z.object({
  title: z
    .string({ message: "Title is required" })
    .min(1, { message: "Title is too short" }),
  url: z
    .string({ message: "URL is required" })
    .url({ message: "URL is invalid" }),
  show_text: z.boolean(),
});

type FormFields = z.infer<typeof schema>;

export const BookmarkAddModal: React.FC<
  React.FormHTMLAttributes<HTMLFormElement>
> = ({ className, ...props }) => {
  const { createBookmark } = useBookmarks();
  const [isChecked, setChecked] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormFields>({ resolver: zodResolver(schema), mode: "onChange" });

  const onSubmit: SubmitHandler<FormFields> = async (data) => {
    createBookmark(data);
    reset();
  };

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className={cn(
        "bg-smoke-900 border-smoke-700 flex w-[360px] flex-col gap-4 rounded-lg border-2 p-4",
        className,
      )}
      {...props}
    >
      <p>Add a Bookmark</p>
      <div className="flex flex-col gap-2">
        <Input
          {...register("title")}
          inputError={errors.title?.message}
          placeholder="Title"
          name="title"
          first
        />
        <Input
          {...register("url")}
          inputError={errors.url?.message}
          placeholder="URL"
          name="url"
        />
        <Checkbox
          {...register("show_text")}
          onChange={() => {
            setChecked(!isChecked);
          }}
          isChecked={isChecked}
        >
          Show Title
        </Checkbox>
        <Button
          text={isSubmitting ? "" : "Save"}
          type="submit"
          className="w-full"
          size="md"
        >
          {isSubmitting ? <Loader width={24} height={24} /> : <></>}
        </Button>
      </div>
    </form>
  );
};
