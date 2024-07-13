import { useEffect } from "react";
import axios from "../utils/requests";
import { useAuth } from "../hooks/useAuth";

export const Secret = () => {
  const { logout } = useAuth();

  useEffect(() => {
    const getFeed = async () => {
      const res = await axios.get("v1/posts");
      console.log(res);
    };

    getFeed();
  }, []);

  const handleLogout = () => {
    logout();
  };

  return (
    <div>
      <button onClick={handleLogout}>Logout</button>
    </div>
  );
};
