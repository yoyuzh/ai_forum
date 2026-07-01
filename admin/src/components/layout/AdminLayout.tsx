import { Outlet } from "react-router-dom";
import TopNav from "./TopNav";
import SideNav from "./SideNav";

/** Admin shell — sticky top nav + fixed left sidebar + offset main content. */
export default function AdminLayout() {
  return (
    <div className="min-h-screen bg-cohere-background">
      <TopNav />
      <SideNav />
      <main className="min-h-[calc(100vh-64px)] md:pl-64">
        <Outlet />
      </main>
    </div>
  );
}
