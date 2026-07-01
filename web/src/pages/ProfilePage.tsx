import { useMemo, useState, useEffect, FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useUserStore } from "../stores/useUserStore";
import { api } from "../api/client";
import type { UserPreferences } from "../api/types";
import MaterialIcon from "../components/ui/MaterialIcon";
import AlertBar from "../components/ui/AlertBar";
import SafeMarkdown from "../components/ui/SafeMarkdown";

const BIO_MAX = 300;
const PRESET_AVATARS = [
  "https://api.dicebear.com/7.x/avataaars/svg?seed=dev1",
  "https://api.dicebear.com/7.x/avataaars/svg?seed=research",
  "https://api.dicebear.com/7.x/avataaars/svg?seed=nova",
  "https://api.dicebear.com/7.x/avataaars/svg?seed=forum",
  "https://api.dicebear.com/7.x/bottts/svg?seed=ArchTechLead",
  "https://api.dicebear.com/7.x/bottts/svg?seed=Devil",
];

interface ProfileForm {
  nickname: string;
  avatar: string;
  bio: string;
  preferences: UserPreferences;
}

function formatJoinedAt(iso: string): string {
  try {
    const d = new Date(iso);
    if (Number.isNaN(d.getTime())) return "—";
    return `${d.getFullYear()}年${d.getMonth() + 1}月加入`;
  } catch {
    return "—";
  }
}

function formatStat(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
  return String(n);
}

export default function ProfilePage() {
  const navigate = useNavigate();
  const currentUser = useUserStore((s) => s.currentUser);
  const updateCurrentUser = useUserStore((s) => s.updateCurrentUser);
  const clearAuthed = useUserStore((s) => s.clearAuthed);

  // Aggregated stats — recomputed from posts/comments, never persisted.
  const { data: stats } = useQuery({
    queryKey: ["user-stats", currentUser?.username],
    queryFn: () => api.user.getStats(currentUser!.username),
    enabled: Boolean(currentUser?.username),
  });

  const [form, setForm] = useState<ProfileForm | null>(null);
  const [avatarUrlDraft, setAvatarUrlDraft] = useState("");
  const [bioError, setBioError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [loggingOut, setLoggingOut] = useState(false);

  // Seed the form whenever the underlying profile changes (e.g. after login).
  useEffect(() => {
    if (currentUser) {
      setForm({
        nickname: currentUser.nickname,
        avatar: currentUser.avatar,
        bio: currentUser.bio,
        preferences: { ...currentUser.preferences },
      });
      setAvatarUrlDraft("");
    }
  }, [currentUser]);

  // isDirty must be computed before any early return, otherwise the hook
  // order changes between renders (Rules of Hooks). Guard nulls internally.
  const isDirty = useMemo(() => {
    if (!currentUser || !form) return false;
    return (
      form.nickname !== currentUser.nickname ||
      form.avatar !== currentUser.avatar ||
      form.bio !== currentUser.bio ||
      form.preferences.aiReplyNotifications !== currentUser.preferences.aiReplyNotifications ||
      form.preferences.liveActivity !== currentUser.preferences.liveActivity ||
      form.preferences.themePreference !== currentUser.preferences.themePreference
    );
  }, [form, currentUser]);

  // Guard: if there is no logged-in user, surface an inline prompt instead of
  // crashing. Real auth lives on the backend; this is a frontend-only nudge.
  if (!currentUser || !form) {
    return (
      <main className="mx-auto flex w-full max-w-4xl flex-grow flex-col items-center justify-center gap-md px-margin-mobile py-section text-center md:px-margin-desktop">
        <MaterialIcon name="lock_person" size={56} className="text-cohere-muted" />
        <h1 className="font-headline-xl text-cohere-primary">尚未登录</h1>
        <p className="font-body-main text-cohere-on-surface-variant">
          登录后即可查看与编辑你的个人资料。
        </p>
        <button type="button" onClick={() => navigate("/login")} className="btn-primary">
          去登录
        </button>
      </main>
    );
  }

  const onLogout = async () => {
    setLoggingOut(true);
    try {
      await api.auth.logout();
      clearAuthed();
      navigate("/login", { replace: true });
    } catch {
      setError("退出登录失败，请稍后重试");
      setLoggingOut(false);
    }
  };

  const setBio = (value: string) => {
    setForm((p) => (p ? { ...p, bio: value } : p));
    if (value.length > BIO_MAX) {
      setBioError(`简介不能超过 ${BIO_MAX} 字，当前 ${value.length} 字`);
    } else {
      setBioError(null);
    }
  };

  const togglePreference = (key: keyof UserPreferences) => {
    setForm((p) => {
      if (!p) return p;
      const current = p.preferences[key];
      return {
        ...p,
        preferences: { ...p.preferences, [key]: !current },
      };
    });
  };

  const onReset = () => {
    setForm({
      nickname: currentUser.nickname,
      avatar: currentUser.avatar,
      bio: currentUser.bio,
      preferences: { ...currentUser.preferences },
    });
    setAvatarUrlDraft("");
    setBioError(null);
    setSuccess(null);
    setError(null);
  };

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    if (bioError) {
      setError("请先修正简介中的问题再保存");
      return;
    }
    const nickname = form.nickname.trim();
    if (!nickname) {
      setError("昵称不能为空");
      return;
    }

    setSubmitting(true);
    try {
      const updated = await api.user.updateProfile({
        nickname,
        avatar: form.avatar,
        bio: form.bio,
        preferences: form.preferences,
      });
      updateCurrentUser({
        nickname: updated.nickname,
        avatar: updated.avatar,
        bio: updated.bio,
        preferences: updated.preferences,
      });
      setSuccess("资料已保存");
      setAvatarUrlDraft("");
    } catch (err) {
      const message = err instanceof Error ? err.message : "保存失败，请稍后重试";
      setError(message);
    } finally {
      setSubmitting(false);
    }
  };

  const STAT_ITEMS = [
    { label: "发帖", value: stats?.postCount ?? 0 },
    { label: "评论", value: stats?.commentCount ?? 0 },
    { label: "获赞", value: stats?.likeCount ?? 0 },
    { label: "AI回复", value: stats?.aiReplyCount ?? 0 },
  ];

  return (
    <main className="mx-auto w-full max-w-4xl flex-grow px-margin-mobile py-section md:px-margin-desktop animate-reveal-up">
      {/* User header & stats board */}
      <section className="flex flex-col gap-xl border-b border-cohere-hairline py-lg md:flex-row md:items-start">
        <div className="flex flex-grow flex-col items-center gap-lg md:flex-row md:items-start">
          <div className="relative shrink-0">
            <img
              src={currentUser.avatar}
              alt={currentUser.nickname || currentUser.username}
              width={96}
              height={96}
              className="h-24 w-24 rounded-full border border-cohere-hairline object-cover shadow-sm"
            />
            <a
              href="#avatar-edit"
              className="absolute bottom-0 right-0 flex h-6 w-6 items-center justify-center rounded-full bg-cohere-primary text-cohere-on-primary transition-colors hover:bg-cohere-ink focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
              aria-label="修改头像"
            >
              <MaterialIcon name="edit" size={12} />
            </a>
          </div>
          <div className="flex max-w-xl flex-col gap-2 text-center md:text-left">
            <div>
              <h1 className="font-headline-lg text-[24px] font-medium tracking-tight text-cohere-primary">
                {currentUser.nickname || currentUser.username}
              </h1>
              <p className="font-label-mono text-micro text-cohere-muted">
                UID: {currentUser.uid} • {formatJoinedAt(currentUser.joinedAt)}
              </p>
            </div>
            {currentUser.bio ? (
              <SafeMarkdown
                content={currentUser.bio}
                className="font-body-main text-caption leading-relaxed text-cohere-on-surface-variant"
              />
            ) : (
              <p className="font-body-main text-caption leading-relaxed text-cohere-muted italic">
                这位研究员还没有填写简介。
              </p>
            )}
          </div>
        </div>

        {/* Stats — wraps on mobile, never overflows. */}
        <div className="flex flex-wrap items-center justify-center gap-lg pt-md md:flex-row md:gap-xl md:pt-0">
          {STAT_ITEMS.map((item, idx) => (
            <div key={item.label} className="flex items-center gap-lg md:gap-xl">
              {idx > 0 && <span className="hidden h-8 w-px bg-cohere-hairline md:block" aria-hidden="true" />}
              <div className="flex flex-col items-center md:items-start">
                <span className="font-headline-lg text-[20px] font-medium text-cohere-primary">
                  {formatStat(item.value)}
                </span>
                <span className="font-micro uppercase tracking-wider text-cohere-muted">
                  {item.label}
                </span>
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* Settings form */}
      <section className="mt-section rounded-ai border border-cohere-hairline bg-cohere-surface-lowest p-lg">
        <div className="mb-xl flex flex-wrap items-end justify-between gap-sm border-b border-cohere-hairline pb-md">
          <div>
            <h2 className="font-headline-lg font-light tracking-wide text-cohere-primary">账号设置</h2>
            <p className="mt-2 font-caption text-cohere-muted">管理您的个人资料和偏好设置。</p>
          </div>
          {currentUser.emailVerified && (
            <span className="rounded-xs bg-cohere-success px-sm py-xxs font-micro font-bold uppercase tracking-wider text-cohere-deep-green">
              邮箱已验证
            </span>
          )}
        </div>

        {success && (
          <div className="mb-lg">
            <AlertBar tone="success" message={success} duration={3000} onClose={() => setSuccess(null)} />
          </div>
        )}
        {error && (
          <div className="mb-lg">
            <AlertBar tone="error" message={error} onClose={() => setError(null)} />
          </div>
        )}

        <form onSubmit={onSubmit} className="flex max-w-2xl flex-col gap-xl" noValidate>
          {/* Avatar edit */}
          <div id="avatar-edit" className="flex flex-col gap-md sm:flex-row sm:items-center">
            <img
              src={form.avatar}
              alt="当前头像预览"
              width={64}
              height={64}
              className="h-16 w-16 rounded-full border border-cohere-hairline object-cover"
            />
            <div className="flex flex-col gap-xs">
              <span className="font-caption text-cohere-muted">从预设头像中选择，或粘贴图片 URL</span>
              <div className="flex flex-wrap gap-xs">
                {PRESET_AVATARS.map((url) => (
                  <button
                    key={url}
                    type="button"
                    onClick={() => {
                      setForm((p) => (p ? { ...p, avatar: url } : p));
                      setAvatarUrlDraft("");
                    }}
                    aria-label="选择该预设头像"
                    aria-pressed={form.avatar === url}
                    className={`h-9 w-9 rounded-full border transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue ${
                      form.avatar === url
                        ? "border-cohere-secondary ring-2 ring-cohere-secondary/40"
                        : "border-cohere-hairline hover:border-cohere-secondary"
                    }`}
                  >
                    <img src={url} alt="" className="h-full w-full rounded-full object-cover" />
                  </button>
                ))}
              </div>
              <div className="flex items-center gap-sm">
                <input
                  id="profile-avatar-url"
                  name="avatar-url"
                  type="url"
                  value={avatarUrlDraft}
                  onChange={(e) => setAvatarUrlDraft(e.target.value)}
                  placeholder="https://…/avatar.png"
                  className="flex-1 rounded-sm border border-cohere-hairline bg-cohere-surface-lowest px-md py-sm font-body-main text-cohere-primary placeholder:text-cohere-muted focus:border-cohere-secondary focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
                />
                <button
                  type="button"
                  onClick={() => {
                    if (avatarUrlDraft.trim()) {
                      setForm((p) => (p ? { ...p, avatar: avatarUrlDraft.trim() } : p));
                    }
                  }}
                  className="btn-pill-outline py-sm"
                >
                  应用
                </button>
              </div>
              <p className="font-micro text-cohere-muted">支持 JPG, GIF 或 PNG。最大 2MB。</p>
            </div>
          </div>

          {/* Input grid */}
          <div className="grid grid-cols-1 gap-lg md:grid-cols-2">
            <div className="flex flex-col gap-sm">
              <label htmlFor="profile-nickname" className="font-caption text-cohere-muted">
                昵称
              </label>
              <input
                id="profile-nickname"
                name="nickname"
                type="text"
                autoComplete="nickname"
                value={form.nickname}
                onChange={(e) => setForm((p) => (p ? { ...p, nickname: e.target.value } : p))}
                className="rounded-sm border border-cohere-hairline bg-cohere-surface-lowest px-md py-sm font-body-main text-cohere-primary focus:border-cohere-form-focus focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
              />
            </div>
            <div className="flex flex-col gap-sm">
              <label htmlFor="profile-email" className="font-caption text-cohere-muted">
                邮箱地址
              </label>
              <div className="relative">
                <input
                  id="profile-email"
                  name="email"
                  type="email"
                  value={currentUser.email}
                  disabled
                  readOnly
                  className="w-full cursor-not-allowed rounded-sm border border-cohere-hairline bg-cohere-surface-low px-md py-sm font-body-main text-cohere-muted"
                />
                {currentUser.emailVerified && (
                  <span className="absolute right-md top-1/2 -translate-y-1/2 rounded-xs bg-cohere-success px-sm py-xxs font-micro font-bold uppercase tracking-wider text-cohere-deep-green">
                    已验证
                  </span>
                )}
              </div>
              <p className="font-micro text-cohere-muted">邮箱为账户凭证，如需修改请联系管理员。</p>
            </div>
          </div>

          {/* Bio */}
          <div className="flex flex-col gap-sm">
            <label htmlFor="profile-bio" className="font-caption text-cohere-muted">
              个人简介
            </label>
            <textarea
              id="profile-bio"
              name="bio"
              rows={4}
              value={form.bio}
              onChange={(e) => setBio(e.target.value)}
              maxLength={BIO_MAX + 50}
              className="resize-none rounded-sm border border-cohere-hairline bg-cohere-surface-lowest px-md py-sm font-body-main text-cohere-primary focus:border-cohere-form-focus focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
            />
            <div className="flex items-center justify-between">
              <span className={`font-micro ${bioError ? "text-cohere-error" : "text-cohere-muted"}`}>
                {bioError ?? "支持纯文本与基础 Markdown。"}
              </span>
              <span className={`font-micro ${form.bio.length > BIO_MAX ? "text-cohere-error" : "text-cohere-muted"}`}>
                {form.bio.length} / {BIO_MAX} 字符
              </span>
            </div>
          </div>

          {/* Preferences */}
          <fieldset className="flex flex-col gap-md rounded-sm border border-cohere-hairline bg-cohere-surface-low p-md">
            <legend className="px-xs font-label-mono-bold text-cohere-primary">偏好设置</legend>

            <label className="flex cursor-pointer items-center justify-between gap-md">
              <span className="flex flex-col">
                <span className="font-body-main text-cohere-on-surface">接收 AI 回复通知</span>
                <span className="font-micro text-cohere-muted">AI 回复你的帖子或评论时提醒你。</span>
              </span>
              <input
                type="checkbox"
                checked={form.preferences.aiReplyNotifications}
                onChange={() => togglePreference("aiReplyNotifications")}
                className="h-4 w-4 rounded border-cohere-hairline text-cohere-primary focus:ring-cohere-primary"
              />
            </label>

            <label className="flex cursor-pointer items-center justify-between gap-md">
              <span className="flex flex-col">
                <span className="font-body-main text-cohere-on-surface">开启实时动态</span>
                <span className="font-micro text-cohere-muted">展示 AI 实时活动与 SSE 在线状态。</span>
              </span>
              <input
                type="checkbox"
                checked={form.preferences.liveActivity}
                onChange={() => togglePreference("liveActivity")}
                className="h-4 w-4 rounded border-cohere-hairline text-cohere-primary focus:ring-cohere-primary"
              />
            </label>

            <div className="flex flex-col gap-xs">
              <label htmlFor="profile-theme" className="font-body-main text-cohere-on-surface">
                主题偏好
              </label>
              <select
                id="profile-theme"
                name="theme"
                value={form.preferences.themePreference}
                onChange={(e) =>
                  setForm((p) =>
                    p
                      ? {
                          ...p,
                          preferences: {
                            ...p.preferences,
                            themePreference: e.target.value as UserPreferences["themePreference"],
                          },
                        }
                      : p,
                  )
                }
                className="w-full rounded-sm border border-cohere-hairline bg-cohere-surface-lowest px-md py-sm font-body-main text-cohere-primary focus:border-cohere-form-focus focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue"
              >
                <option value="system">跟随系统</option>
                <option value="light">浅色</option>
                <option value="dark">深色</option>
              </select>
              <p className="font-micro text-cohere-muted">主题切换为占位功能，暂未生效。</p>
            </div>
          </fieldset>

          {/* Actions */}
          <div className="flex flex-col items-center justify-between gap-md border-t border-cohere-hairline pt-lg sm:flex-row">
            <button
              type="button"
              onClick={onLogout}
              disabled={loggingOut}
              className="flex items-center gap-sm rounded-lg px-md py-sm font-body-main text-cohere-error transition-colors hover:bg-cohere-error-container hover:text-cohere-on-error-container focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue disabled:opacity-50"
            >
              <MaterialIcon name="logout" size={20} />
              {loggingOut ? "退出中…" : "退出登录"}
            </button>
            <div className="flex w-full gap-md sm:w-auto">
              <button
                type="button"
                onClick={onReset}
                disabled={!isDirty || submitting}
                className="flex-1 rounded-pill px-lg py-sm font-body-main text-cohere-on-surface-variant transition-colors hover:bg-cohere-surface-low focus:outline-none focus-visible:ring-2 focus-visible:ring-cohere-focus-blue disabled:opacity-40 sm:flex-none"
              >
                取消
              </button>
              <button
                type="submit"
                disabled={submitting || !isDirty || Boolean(bioError)}
                className="btn-primary flex-1 px-lg py-sm sm:flex-none"
              >
                {submitting ? "保存中…" : "保存更改"}
              </button>
            </div>
          </div>
        </form>
      </section>
    </main>
  );
}
