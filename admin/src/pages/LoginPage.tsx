import { useState } from "react";
import { Button, Input, App as AntdApp } from "antd";
import { useNavigate } from "react-router-dom";
import { adminApi } from "../api/client";

export default function LoginPage() {
  const navigate = useNavigate();
  const { message } = AntdApp.useApp();
  const [username, setUsername] = useState("admin");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit() {
    setLoading(true);
    try {
      await adminApi.auth.login(username, password);
      navigate("/", { replace: true });
    } catch {
      message.error("登录失败");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="grid min-h-screen place-items-center bg-cohere-background px-margin-mobile">
      <section className="w-full max-w-sm rounded-lg border border-cohere-hairline bg-cohere-surface-lowest p-lg">
        <h1 className="font-feature-title text-cohere-primary">AI Forum Admin</h1>
        <div className="mt-lg flex flex-col gap-md">
          <label className="font-label-mono text-cohere-muted" htmlFor="admin-username">用户名</label>
          <Input id="admin-username" aria-label="用户名" value={username} onChange={(e) => setUsername(e.target.value)} />
          <label className="font-label-mono text-cohere-muted" htmlFor="admin-password">密码</label>
          <Input.Password id="admin-password" aria-label="密码" value={password} onChange={(e) => setPassword(e.target.value)} onPressEnter={submit} />
          <Button type="primary" loading={loading} onClick={submit}>登录</Button>
        </div>
      </section>
    </main>
  );
}
