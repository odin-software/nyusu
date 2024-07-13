import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App.tsx";
import "./index.css";

import { BrowserRouter } from "react-router-dom";
import { useRoutes } from "react-router-dom";
import { AuthProvider } from "./hooks/useAuth.tsx";
import { ProtectedRoute } from "./components/ProtectedRoute.tsx";
import { Secret } from "./pages/Secret.tsx";
import { Login } from "./pages/Login.tsx";
// ...

function Root() {
  const routes = useRoutes([
    {
      path: "/",
      element: <App />,
    },
    {
      path: "/login",
      element: <Login />,
    },
    {
      path: "/posts",
      element: (
        <ProtectedRoute>
          <Secret />
        </ProtectedRoute>
      ),
    },
  ]);
  return routes;
}

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <BrowserRouter>
      <AuthProvider>
        <Root />
      </AuthProvider>
    </BrowserRouter>
  </React.StrictMode>
);
