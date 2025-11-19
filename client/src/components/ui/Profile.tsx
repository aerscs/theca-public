import { useRef, useState } from "react";
import { useAuth } from "@/hooks/useAuth";
import { useOutsideClick } from "@/hooks/useOutsideClick";
import { useNavigate } from "react-router";
import { Button } from "./Button";
import { LogOut, User } from "lucide-react";

export const Profile = () => {
  const { handleLogout, currentUser } = useAuth();
  const navigate = useNavigate();

  const [logoutClicked, setLogoutClicked] = useState(false);

  const buttonRef = useRef<HTMLButtonElement>(null);

  const handleClick = () => {
    if (!logoutClicked) {
      setLogoutClicked(true);
    } else {
      handleLogout();
      navigate("/login");
    }
  };

  useOutsideClick(buttonRef, () => {
    setLogoutClicked(false);
  });

  return (
    <Button
      ref={buttonRef}
      text={logoutClicked ? "Log out?" : currentUser!.username}
      onClick={handleClick}
      reversed
      icon={logoutClicked ? LogOut : User}
      className="max-w-52 items-start pr-2"
    />
  );
};
