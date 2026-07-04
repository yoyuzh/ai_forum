import { Outlet } from "react-router-dom";
import TopNav from "./TopNav";
import SideNav from "./SideNav";

/** Admin shell — fixed console chrome with scrollable content. */
export default function AdminLayout() {
  return (
    <div className="min-h-screen bg-cohere-surface">
      <TopNav />
      <SideNav />
      <main className="min-h-[calc(100vh-64px)] border-l border-cohere-hairline bg-cohere-surface md:ml-64">
        <Outlet />
      </main>
    </div>
  );
}
