import { createContext, useContext, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { useLocalStorage } from "./useLocalStorage";
import axios from "axios";

type User = {
  id: number;
  name: string;
  email: string;
  created_at: Date;
  updated_at: Date;
};

type LoginData = {
  email: string;
  password: string;
};

type AuthValue = {
  user: User | null;
  login: (data: LoginData) => Promise<boolean>;
  logout: () => void;
};

const initialContext: AuthValue = {
  user: null,
  login: async (_) => {
    console.log("not implemented yet");
    return false;
  },
  logout: () => {},
};

const AuthContext = createContext<AuthValue>(initialContext);

export const AuthProvider = ({ children }: { children: React.ReactNode }) => {
  const [user, setUser] = useLocalStorage<User | null>("user", null);
  const navigate = useNavigate();

  const login = async (data: LoginData): Promise<boolean> => {
    const res = await fetch("http://localhost:8888/v1/users/login", {
      method: "POST",
      body: JSON.stringify(data),
    });
    if (!res.ok) {
      return false;
    }
    const resData = await res.json();
    console.log(resData);
    const userResponse = await axios("http://localhost:8888/v1/users", {
      method: "GET",
      headers: {
        "Content-type": "application/json",
        Authorization: `Bearer ${resData.token}`,
      },
    });
    if (userResponse.status !== 200) {
      return false;
    }
    const userData = (await resData.json()) as User;
    setUser(userData);
    navigate("/");
    return true;
  };

  const logout = () => {
    setUser(null);
    navigate("/login", { replace: true });
  };

  const value = useMemo(
    () => ({
      user,
      login,
      logout,
    }),
    [user]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

export const useAuth = () => {
  return useContext(AuthContext);
};
