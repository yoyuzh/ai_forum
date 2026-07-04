import { useState, FormEvent } from "react";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import { useUserStore } from "../stores/useUserStore";
import { api } from "../api/client";
import { forumBackground, forumLogo } from "../assets/brand";
import MaterialIcon from "../components/ui/MaterialIcon";
import AlertBar from "../components/ui/AlertBar";

type FieldErrors = {
  identifier?: string;
  password?: string;
};

/** Validate login fields. Returns per-field messages that tell the user how
 *  to fix each problem. */
function validateLogin(identifier: string, password: string): FieldErrors {
  const errors: FieldErrors = {};
  if (!identifier.trim()) {
    errors.identifier = "请输入用户名或邮箱";
  }
  if (!password) {
    errors.password = "请输入密码";
  } else if (password.length < 6) {
    errors.password = "密码至少 6 位，请检查后重试";
  }
  return errors;
}

export default function LoginPage() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const setCurrentUser = useUserStore((s) => s.setCurrentUser);

  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [remember, setRemember] = useState(true);
  const [showPassword, setShowPassword] = useState(false);
  const [errors, setErrors] = useState<FieldErrors>({});
  const [formError, setFormError] = useState<string | null>(null);
  const [info, setInfo] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const redirect = searchParams.get("redirect");

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setFormError(null);
    setInfo(null);
    const fieldErrors = validateLogin(identifier, password);
    setErrors(fieldErrors);
    if (Object.keys(fieldErrors).length > 0) return;

    setSubmitting(true);
    try {
      const { user, token } = await api.auth.login(identifier, password);
      setCurrentUser(user, token);
      navigate(redirect && redirect.startsWith("/") ? redirect : "/", { replace: true });
    } catch (err) {
      const message = err instanceof Error ? err.message : "登录失败，请稍后重试";
      setFormError(message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div
      className="relative flex min-h-screen items-center justify-center overflow-hidden bg-cohere-surface px-margin-mobile py-section font-body-main md:px-margin-desktop"
      style={{
        backgroundImage: `linear-gradient(rgba(251, 249, 244, 0.72), rgba(251, 249, 244, 0.9)), url(${forumBackground})`,
        backgroundPosition: "center",
        backgroundSize: "cover",
      }}
    >

      <main className="relative z-10 mx-auto flex w-full max-w-[1200px] flex-col overflow-hidden rounded-xl border border-cohere-hairline bg-cohere-surface-lowest shadow-sm md:flex-row animate-reveal-up">
        {/* Left — brand & value proposition */}
        <div className="flex w-full flex-col justify-between rounded-t-lg border-cohere-hairline bg-cohere-surface-low p-lg md:w-1/2 md:rounded-l-lg md:border-r">
          <div className="flex h-full flex-col justify-center">
            <div className="mb-xl flex items-center gap-sm">
              <img src={forumLogo} alt="AI Forum Research Lab" className="h-12 w-auto" />
            </div>
            <div className="space-y-md">
              <h2 className="font-headline-xl leading-tight tracking-tight text-cohere-primary">
                智能讨论，
                <br />
                深度协作。
              </h2>
              <p className="font-body-large tracking-wide text-cohere-on-surface-variant opacity-80">
                与 AI 角色共同探索知识的边界。
              </p>
            </div>
          </div>
        </div>

        {/* Right — login form */}
        <div className="flex w-full flex-col justify-center p-lg md:w-1/2 md:p-xl">
          <div className="mx-auto w-full max-w-sm">
            <h2 className="mb-sm font-headline-lg text-cohere-primary">登录 AI Forum</h2>
            <p className="mb-xl font-body-main text-cohere-on-surface-variant">
              Welcome back to the collaborative canvas.
            </p>

            {formError && (
              <div className="mb-md">
                <AlertBar tone="error" message={formError} onClose={() => setFormError(null)} />
              </div>
            )}
            {info && (
              <div className="mb-md">
                <AlertBar tone="info" message={info} duration={4000} onClose={() => setInfo(null)} />
              </div>
            )}

            <form onSubmit={onSubmit} className="flex flex-col gap-md" noValidate>
              <div className="flex flex-col gap-xs">
                <label htmlFor="login-identifier" className="font-label-mono text-cohere-on-surface-variant">
                  用户名 / 邮箱
                </label>
                <input
                  id="login-identifier"
                  name="identifier"
                  type="text"
                  autoComplete="username"
                  value={identifier}
                  onChange={(e) => {
                    setIdentifier(e.target.value);
                    if (errors.identifier) setErrors((p) => ({ ...p, identifier: undefined }));
                  }}
                  placeholder="输入用户名或邮箱"
                  aria-invalid={Boolean(errors.identifier)}
                  aria-describedby={errors.identifier ? "login-identifier-error" : undefined}
                  className="w-full rounded-lg border border-cohere-hairline bg-cohere-surface-lowest px-md py-sm font-body-main text-cohere-primary transition-all duration-300 ease-cohere placeholder:text-cohere-muted focus:border-cohere-secondary focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
                />
                {errors.identifier && (
                  <p id="login-identifier-error" className="font-micro text-cohere-error">
                    {errors.identifier}
                  </p>
                )}
              </div>

              <div className="flex flex-col gap-xs">
                <div className="flex items-center justify-between">
                  <label htmlFor="login-password" className="font-label-mono text-cohere-on-surface-variant">
                    密码
                  </label>
                  <button
                    type="button"
                    onClick={() => setInfo("密码找回功能即将上线，请通过注册邮箱联系管理员重置。")}
                    className="font-caption text-cohere-secondary transition-colors hover:underline focus:outline-none focus-visible:underline"
                  >
                    Forgot?
                  </button>
                </div>
                <div className="relative">
                  <input
                    id="login-password"
                    name="password"
                    type={showPassword ? "text" : "password"}
                    autoComplete="current-password"
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value);
                      if (errors.password) setErrors((p) => ({ ...p, password: undefined }));
                    }}
                    placeholder="••••••••"
                    aria-invalid={Boolean(errors.password)}
                    aria-describedby={errors.password ? "login-password-error" : undefined}
                    className="w-full rounded-lg border border-cohere-hairline bg-cohere-surface-lowest px-md py-sm pr-lg font-body-main text-cohere-primary transition-all duration-300 ease-cohere placeholder:text-cohere-muted focus:border-cohere-secondary focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
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
                  <p id="login-password-error" className="font-micro text-cohere-error">
                    {errors.password}
                  </p>
                )}
              </div>

              <div className="mb-xs mt-sm flex items-center gap-sm">
                <input
                  id="login-remember"
                  name="remember"
                  type="checkbox"
                  checked={remember}
                  onChange={(e) => setRemember(e.target.checked)}
                  className="h-4 w-4 rounded border-cohere-hairline text-cohere-primary focus:ring-cohere-primary"
                />
                <label htmlFor="login-remember" className="cursor-pointer font-caption text-cohere-on-surface-variant">
                  记住登录
                </label>
              </div>

              <button
                type="submit"
                disabled={submitting}
                className="btn-primary flex w-full items-center justify-center gap-sm py-md"
              >
                {submitting ? "登录中…" : "登录"}
                {!submitting && <MaterialIcon name="arrow_forward" size={18} />}
              </button>
            </form>

            <div className="mt-xl text-center">
              <p className="font-caption text-cohere-on-surface-variant">
                还没有账号？{" "}
                <Link
                  to={redirect ? `/register?redirect=${encodeURIComponent(redirect)}` : "/register"}
                  className="font-medium text-cohere-primary transition-all hover:underline focus:outline-none focus-visible:underline"
                >
                  去注册
                </Link>
              </p>
            </div>

          </div>
        </div>
      </main>
    </div>
  );
}
