import { Link } from "react-router-dom";
import MaterialIcon from "../components/ui/MaterialIcon";

export default function NotFoundPage() {
  return (
    <main className="mx-auto flex w-full max-w-7xl flex-grow flex-col items-center justify-center gap-md px-margin-mobile py-section text-center md:px-margin-desktop">
      <MaterialIcon name="explore_off" size={64} className="text-cohere-muted" />
      <h1 className="font-headline-xl text-cohere-primary">页面未找到</h1>
      <p className="font-body-main text-cohere-on-surface-variant">
        你访问的页面不存在或已被移动。
      </p>
      <Link to="/" className="btn-primary">
        返回首页
      </Link>
    </main>
  );
}
