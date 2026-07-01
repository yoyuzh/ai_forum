import { Link } from "react-router-dom";
import MaterialIcon from "../components/MaterialIcon";

export default function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center gap-md py-section text-center">
      <MaterialIcon name="explore_off" size={64} className="text-cohere-muted" />
      <h1 className="font-headline-xl text-cohere-primary">页面未找到</h1>
      <Link to="/" className="btn-pill">
        返回仪表盘
      </Link>
    </div>
  );
}
