import { useAuth } from "@/hooks/useAuth";
import { Outlet, Navigate } from "react-router";

export const ProtectedRoute: React.FC = () => {
  const { currentUser, isLoading } = useAuth();

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (!isLoading && !currentUser) return <Navigate to="/login" replace />;

  return <Outlet />;
};
