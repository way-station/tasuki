import { RouterProvider } from "@tanstack/react-router";
import { createAppRouter } from "./router.tsx";

const router = createAppRouter();

export function App() {
  return <RouterProvider router={router} />;
}
