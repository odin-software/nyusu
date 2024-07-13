import axios from "axios";

const a = axios.create({
  baseURL: "http://localhost:8888/",
  headers: {
    "Content-Type": "application/json",
  },
});

a.interceptors.request.use(
  (request) => {
    const token = localStorage.getItem("token");
    if (!token) {
      return request;
    }
    request.headers.Authorization = `Bearer ${JSON.parse(token)}`;
    return request;
  },
  (error) => error
);

export default a;
