import { Input } from "@/components/Form/Input";
import { Checkbox } from "../Form/Checkbox";
import { useState } from "react";
import { z } from "zod";
import { useForm, type SubmitHandler } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "@/components/ui/Button";
import { cn } from "@/lib/utils";
import { useBookmarks } from "@/hooks/useBookmarks";
import { Trash2 } from "lucide-react";
import { Loader } from "../ui/Loader";

interface Props extends React.FormHTMLAttributes<HTMLFormElement> {
  bookmarkId: number;
  title: string;
  url: string;
  show_text: boolean;
}

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

export const BookmarkEditModal: React.FC<Props> = ({
  bookmarkId,
  className,
  title,
  url,
  show_text,
  ...props
}) => {
  const { updateBookmark, deleteBookmark } = useBookmarks();

  const [titleState, setTitleState] = useState(title);
  const [urlState, setUrlState] = useState(url);
  const [isChecked, setChecked] = useState(show_text);
  const [isDeleting, setIsDeleting] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<FormFields>({ resolver: zodResolver(schema), mode: "onChange" });

  const onSubmit: SubmitHandler<FormFields> = async (data) => {
    updateBookmark({ ...data, id: bookmarkId });
    reset();
  };

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      className={cn(
        "bg-smoke-900 border-smoke-700 z-50 flex w-[360px] flex-col gap-4 rounded-lg border-2 p-4",
        className,
      )}
      {...props}
    >
      <p>Edit Bookmark</p>
      <div className="flex flex-col gap-2">
        <Input
          {...register("title")}
          value={titleState}
          onChange={(e) => setTitleState(e.target.value)}
          inputError={errors.title?.message}
          placeholder="Title"
          name="title"
          first
        />
        <Input
          {...register("url")}
          value={urlState}
          onChange={(e) => setUrlState(e.target.value)}
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
        {isDeleting ? (
          <Button
            text={`Delete ${title}?`}
            type="button"
            onClick={() => deleteBookmark(bookmarkId)}
            className="w-full"
            data-danger
            size="md"
          />
        ) : (
          <div className="flex gap-2">
            <Button
              onClick={() => setIsDeleting(true)}
              type="button"
              size="md"
              data-danger
              icon={Trash2}
              tabIndex={-1}
            />
            <Button
              text={isSubmitting ? "" : "Save"}
              type="submit"
              className="w-full"
              size="md"
            >
              {isSubmitting ? <Loader width={24} height={24} /> : <></>}
            </Button>
          </div>
        )}
      </div>
    </form>
  );
};
