import { Header } from "@/components/Header";
import { Outlet } from "react-router";

export const Layout = () => {
  return (
    <div className="mx-auto flex h-dvh max-w-[920px] flex-col items-center p-2 sm:px-5">
      <Header />
      <Outlet />
    </div>
  );
};
