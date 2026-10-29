import { BrowserRouter } from "react-router";
import { AppRoutes } from "./routes.tsx";

export function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  );
}
