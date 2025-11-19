import { useEffect, useState } from "react";
import { AuthContext, type Credentials, type User } from "@/hooks/useAuth";
import { api } from "@/api/axiosInstance";

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [authToken, setAuthToken] = useState<string | null>(null);
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  // const [isError, setIsError] = useState<string | null>(null);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const { data } = await api.get("/v1/api/user/me");

        setAuthToken(data.data.access_token);
        setCurrentUser(data.data);
      } catch {
        setAuthToken(null);
        setCurrentUser(null);
      } finally {
        setIsLoading(false);
      }
    };

    fetchUser();
  }, []);

  const handleLogout = async () => {
    await api.delete("/v1/api/logout");
    setAuthToken(null);
    setCurrentUser(null);
  };

  const handleLogin = async (credentials: Credentials) => {
    const { data } = await api.post("/v1/login", credentials);

    setAuthToken(data.data.access_token);
    setCurrentUser(data.data.user);

    api.defaults.headers.common["Authorization"] =
      `Bearer ${data.data.access_token}`;
  };

  return (
    <AuthContext.Provider
      value={{
        authToken,
        currentUser,
        isLoading,
        // isError,
        handleLogin,
        handleLogout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
