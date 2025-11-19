import { createContext, useContext } from "react";

export interface User {
  id: number;
  email: string;
  username: string;
  is_premium: boolean;
}

export interface Credentials {
  username: string;
  password: string;
}

interface AuthContextType {
  authToken: string | null;
  isLoading: boolean;
  // isError: string | null;
  currentUser: User | null;
  handleLogin: (credentials: Credentials) => Promise<void>;
  handleLogout: () => Promise<void>;
}

export const AuthContext = createContext<AuthContextType | undefined>(
  undefined,
);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};
