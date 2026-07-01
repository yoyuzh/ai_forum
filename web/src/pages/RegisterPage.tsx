import { useState, FormEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useUserStore } from "../stores/useUserStore";
import { api } from "../api/client";
import MaterialIcon from "../components/ui/MaterialIcon";
import AlertBar from "../components/ui/AlertBar";

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
const USERNAME_RE = /^[A-Za-z0-9_]+$/;

type RegisterForm = {
  username: string;
  email: string;
  nickname: string;
  password: string;
  confirmPassword: string;
};
type FieldErrors = Partial<Record<keyof RegisterForm, string>>;

function validateRegister(form: RegisterForm): FieldErrors {
  const errors: FieldErrors = {};
  if (!form.username.trim()) {
    errors.username = "请输入用户名";
  } else if (form.username.trim().length < 3) {
    errors.username = "用户名至少 3 个字符";
  } else if (!USERNAME_RE.test(form.username.trim())) {
    errors.username = "用户名只能包含字母、数字和下划线";
  }
  if (!form.email.trim()) {
    errors.email = "请输入邮箱";
  } else if (!EMAIL_RE.test(form.email.trim())) {
    errors.email = "邮箱格式不正确，请检查后重试";
  }
  if (!form.password) {
    errors.password = "请输入密码";
  } else if (form.password.length < 8) {
    errors.password = "密码至少 8 位，建议混合字母与数字";
  }
  if (!form.confirmPassword) {
    errors.confirmPassword = "请再次输入密码";
  } else if (form.confirmPassword !== form.password) {
    errors.confirmPassword = "两次输入的密码不一致，请重新输入";
  }
  return errors;
}

export default function RegisterPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const setCurrentUser = useUserStore((s) => s.setCurrentUser);

  const [form, setForm] = useState<RegisterForm>({
    username: "",
    email: "",
    nickname: "",
    password: "",
    confirmPassword: "",
  });
  const [showPassword, setShowPassword] = useState(false);
  const [errors, setErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const redirect = searchParams.get("redirect");

  const setField = (key: keyof RegisterForm, value: string) => {
    setForm((p) => ({ ...p, [key]: value }));
    if (errors[key]) setErrors((p) => ({ ...p, [key]: undefined }));
  };

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setFormError(null);
    const fieldErrors = validateRegister(form);
    setErrors(fieldErrors);
    if (Object.keys(fieldErrors).length > 0) return;

    setSubmitting(true);
    try {
      const { user } = await api.auth.register({
        username: form.username,
        nickname: form.nickname,
        email: form.email,
        password: form.password,
      });
      setCurrentUser(user);
      navigate(redirect && redirect.startsWith("/") ? redirect : "/profile", { replace: true });
    } catch (err) {
      const message = err instanceof Error ? err.message : "注册失败，请稍后重试";
      setFormError(message);
    } finally {
      setSubmitting(false);
    }
  };

  const inputClass =
    "w-full rounded-lg border border-cohere-hairline bg-cohere-surface-lowest px-md py-sm font-body-main text-cohere-primary transition-all duration-300 ease-cohere placeholder:text-cohere-muted focus:border-cohere-secondary focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue";

  return (
    <div className="flex min-h-screen flex-col bg-cohere-surface font-body-main md:flex-row">
      {/* Left — brand panel (hidden on mobile) */}
      <aside className="relative hidden flex-col justify-between overflow-hidden border-cohere-hairline bg-cohere-surface-low p-section md:flex md:w-1/2 md:border-r">
        <div className="relative z-10">
          <Link
            to="/"
            className="mb-xl inline-flex items-center gap-sm font-headline-lg font-black text-cohere-primary focus:outline-none focus-visible:underline"
          >
            <MaterialIcon name="smart_toy" fill className="text-cohere-primary" />
            AI Forum
          </Link>
          <h1 className="mb-lg font-display-lg text-cohere-primary">
            加入智能
            <br />
            讨论社区
          </h1>
          <p className="max-w-md font-body-large text-cohere-on-surface-variant">
            自定义 AI 交互偏好，参与前沿科技讨论，与全球开发者共同探索人工智能的未来边界。
          </p>
        </div>
        <div className="relative z-10 flex flex-col gap-md">
          <div className="w-fit rounded-lg border border-cohere-hairline bg-cohere-surface p-lg">
            <div className="mb-sm flex items-center gap-md">
              <MaterialIcon name="psychology" className="text-cohere-coral" />
              <span className="font-label-mono-bold">AI_ASSISTANT_READY</span>
            </div>
            <p className="font-caption text-cohere-muted">
              System initialization complete. Awaiting new user parameters.
            </p>
          </div>
        </div>
        {/* Faint geometric backdrop — CSS radial instead of an external image. */}
        <div
          className="pointer-events-none absolute inset-0 z-0 opacity-30 mix-blend-multiply"
          style={{
            backgroundColor: "var(--c-pale-green)",
            backgroundImage:
              "radial-gradient(circle at 20% 30%, var(--c-secondary-container) 0, transparent 40%), radial-gradient(circle at 80% 70%, var(--c-pale-blue) 0, transparent 40%)",
          }}
        />
      </aside>

      {/* Right — register form */}
      <div className="flex flex-1 items-center justify-center bg-cohere-surface px-margin-mobile py-section md:px-margin-desktop animate-reveal-up">
        <div className="w-full max-w-[440px]">
          <div className="mb-xl text-center md:hidden">
            <Link
              to="/"
              className="inline-flex items-center gap-sm font-headline-lg-mobile font-black text-cohere-primary focus:outline-none focus-visible:underline"
            >
              <MaterialIcon name="smart_toy" fill className="text-cohere-primary" />
              AI Forum
            </Link>
          </div>

          <div className="mb-xl">
            <h2 className="mb-xs font-headline-xl text-cohere-primary">注册</h2>
            <p className="font-body-main text-cohere-on-surface-variant">创建您的 AI Forum 账号</p>
          </div>

          {formError && (
            <div className="mb-lg">
              <AlertBar tone="error" message={formError} onClose={() => setFormError(null)} />
            </div>
          )}

          <form onSubmit={onSubmit} className="space-y-lg" noValidate>
            <div className="space-y-sm">
              <label htmlFor="reg-username" className="block font-label-mono-bold text-cohere-primary">
                用户名
              </label>
              <input
                id="reg-username"
                name="username"
                type="text"
                autoComplete="username"
                value={form.username}
                onChange={(e) => setField("username", e.target.value)}
                placeholder="输入用户名"
                aria-invalid={Boolean(errors.username)}
                aria-describedby={errors.username ? "reg-username-error" : undefined}
                className={inputClass}
              />
              {errors.username && (
                <p id="reg-username-error" className="font-micro text-cohere-error">
                  {errors.username}
                </p>
              )}
            </div>

            <div className="space-y-sm">
              <label htmlFor="reg-email" className="block font-label-mono-bold text-cohere-primary">
                邮箱
              </label>
              <input
                id="reg-email"
                name="email"
                type="email"
                autoComplete="email"
                value={form.email}
                onChange={(e) => setField("email", e.target.value)}
                placeholder="name@example.com"
                aria-invalid={Boolean(errors.email)}
                aria-describedby={errors.email ? "reg-email-error" : undefined}
                className={inputClass}
              />
              {errors.email && (
                <p id="reg-email-error" className="font-micro text-cohere-error">
                  {errors.email}
                </p>
              )}
            </div>

            <div className="space-y-sm">
              <label htmlFor="reg-nickname" className="block font-label-mono-bold text-cohere-primary">
                昵称
              </label>
              <input
                id="reg-nickname"
                name="nickname"
                type="text"
                autoComplete="nickname"
                value={form.nickname}
                onChange={(e) => setField("nickname", e.target.value)}
                placeholder="显示在论坛的名称"
                className={inputClass}
              />
              <p className="font-micro text-cohere-muted">留空则使用用户名作为昵称。</p>
            </div>

            <div className="space-y-sm">
              <label htmlFor="reg-password" className="block font-label-mono-bold text-cohere-primary">
                密码
              </label>
              <div className="relative">
                <input
                  id="reg-password"
                  name="password"
                  type={showPassword ? "text" : "password"}
                  autoComplete="new-password"
                  value={form.password}
                  onChange={(e) => setField("password", e.target.value)}
                  placeholder="••••••••"
                  aria-invalid={Boolean(errors.password)}
                  aria-describedby={errors.password ? "reg-password-error" : undefined}
                  className={`${inputClass} pr-lg`}
                />
                <button
                  type="button"
                  onClick={() => setShowPassword((v) => !v)}
                  aria-label={showPassword ? "隐藏密码" : "显示密码"}
                  className="absolute right-md top-1/2 -translate-y-1/2 flex items-center text-cohere-muted transition-colors hover:text-cohere-primary focus:outline-none focus-visible:text-cohere-primary"
                >
                  <MaterialIcon name={showPassword ? "visibility" : "visibility_off"} size={20} />
                </button>
              </div>
              {errors.password && (
                <p id="reg-password-error" className="font-micro text-cohere-error">
                  {errors.password}
                </p>
              )}
            </div>

            <div className="space-y-sm">
              <label
                htmlFor="reg-confirm-password"
                className="block font-label-mono-bold text-cohere-primary"
              >
                确认密码
              </label>
              <input
                id="reg-confirm-password"
                name="confirm-password"
                type={showPassword ? "text" : "password"}
                autoComplete="new-password"
                value={form.confirmPassword}
                onChange={(e) => setField("confirmPassword", e.target.value)}
                placeholder="••••••••"
                aria-invalid={Boolean(errors.confirmPassword)}
                aria-describedby={errors.confirmPassword ? "reg-confirm-error" : undefined}
                className={inputClass}
              />
              {errors.confirmPassword && (
                <p id="reg-confirm-error" className="font-micro text-cohere-error">
                  {errors.confirmPassword}
                </p>
              )}
            </div>

            <button
              type="submit"
              disabled={submitting}
              className="btn-primary mt-xl w-full py-md"
            >
              {submitting ? "注册中…" : "注册"}
            </button>
          </form>

          <div className="mt-lg text-center">
            <p className="font-body-main text-cohere-on-surface-variant">
              已有账号？{" "}
              <Link
                to={redirect ? `/login?redirect=${encodeURIComponent(redirect)}` : "/login"}
                className="font-label-mono-bold text-cohere-primary underline transition-colors hover:text-cohere-coral focus:outline-none focus-visible:text-cohere-coral"
              >
                去登录
              </Link>
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
