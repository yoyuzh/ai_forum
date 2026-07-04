import { Outlet } from "react-router-dom";
import Header from "./Header";
import Footer from "./Footer";

/** App shell shared by every user-facing page. */
export default function AppLayout() {
  return (
    <div className="brand-background flex min-h-screen flex-col bg-cohere-background">
      <Header />
      <Outlet />
      <Footer />
    </div>
  );
}
