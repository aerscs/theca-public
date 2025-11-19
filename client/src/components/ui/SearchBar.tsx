import { useState } from "react";
import { Button } from "@/components/ui/Button";
import { Plus, Search } from "lucide-react";
import { Input } from "@/components/Form/Input";

export const SearchBar: React.FC<
  React.FormHTMLAttributes<HTMLFormElement>
> = () => {
  const [query, setQuery] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (query.trim()) {
      window.open(
        `https://www.google.com/search?q=${encodeURIComponent(query)}`,
        "_blank",
      );
    }
  };

  return (
    <form onSubmit={handleSubmit} className="flex w-[480px] items-center gap-1">
      <Button type="button" icon={Plus} size="md" />
      <Input
        placeholder="Search"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
      />
      <Button type="submit" icon={Search} size="md" />
    </form>
  );
};
