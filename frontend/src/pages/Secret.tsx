import { useEffect, useState } from "react";
import axios from "../utils/requests";
import { useAuth } from "../hooks/useAuth";
import { Post } from "../types";

export const Secret = () => {
  const { logout } = useAuth();
  const [posts, setPosts] = useState<Post[] | null>(null);
  const [url, setUrl] = useState<string>("");

  useEffect(() => {
    const getFeed = async () => {
      const res = await axios.get("v1/posts?pageSize=30");
      console.log(res.data);
      setPosts(res.data);
    };

    getFeed();
  }, []);

  const addFeed = async () => {
    const res = await axios.post("v1/feeds", {
      url,
    });
    console.log(res);
  };

  const handleLogout = () => {
    logout();
  };

  return (
    <div>
      <div
        style={{
          display: "flex",
          flexDirection: "row",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <h2>Nyusu</h2>
        <button onClick={handleLogout}>Logout</button>
      </div>
      <form onSubmit={() => addFeed()}>
        <input
          name="url"
          type="text"
          onChange={(t) => setUrl(t.currentTarget.value)}
        />
        <button type="submit"> Add Feed</button>
      </form>
      <ul
        style={{
          listStyle: "none",
          paddingLeft: "0em",
        }}
      >
        {posts?.map((p) => (
          <li
            key={p.id}
            style={{
              backgroundColor: "black",
              gap: "2px",
              boxShadow: "rgba(17, 12, 46, 0.15) 0px 48px 100px 0px",
              padding: "1em 2em",
              margin: "1em 0em",
              borderRadius: "0.6em",
            }}
          >
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                alignItems: "flex-start",
              }}
            >
              <a
                className="post-title"
                href={p.url}
                target="_blank"
                rel="noopener noreferrer"
              >
                {p.title}
              </a>
              <a
                className="post-author"
                href={p.url}
                target="_blank"
                rel="noopener noreferrer"
              >
                {p.name}
              </a>
            </div>
            <div></div>
          </li>
        ))}
      </ul>
    </div>
  );
};
