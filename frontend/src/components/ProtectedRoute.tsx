import { useAuth } from "../hooks/useAuth";
import { Navigate } from "react-router-dom";

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const auth = useAuth();
  if (!auth.user) {
    return <Navigate to={"/login"} />;
  }
  return children;
}
