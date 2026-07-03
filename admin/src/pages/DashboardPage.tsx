import { useQuery } from "@tanstack/react-query";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { adminApi } from "../api/client";
import StatCard from "../components/StatCard";
import MaterialIcon from "../components/MaterialIcon";

export default function DashboardPage() {
  const { data: stats } = useQuery({ queryKey: ["dashboard", "stats"], queryFn: adminApi.dashboard.stats });
  const { data: trend = [] } = useQuery({
    queryKey: ["dashboard", "trend"],
    queryFn: adminApi.dashboard.weeklyTrend,
  });
  const { data: breakdown } = useQuery({
    queryKey: ["dashboard", "breakdown"],
    queryFn: adminApi.dashboard.taskStatusBreakdown,
  });
  const { data: services = [] } = useQuery({
    queryKey: ["dashboard", "services"],
    queryFn: adminApi.dashboard.services,
  });
  const { data: recentPosts = [] } = useQuery({
    queryKey: ["dashboard", "recentPosts"],
    queryFn: adminApi.dashboard.recentPosts,
  });
  const { data: recentTasks = [] } = useQuery({
    queryKey: ["dashboard", "recentTasks"],
    queryFn: adminApi.dashboard.recentTasks,
  });
  const { data: decisionTimeline = [] } = useQuery({
    queryKey: ["dashboard", "decisionTimeline"],
    queryFn: adminApi.dashboard.decisionTimeline,
  });

  const statusRows = breakdown
    ? [
        { name: "Success", value: breakdown.success },
        { name: "Running", value: breakdown.running },
        { name: "Failed", value: breakdown.failed },
      ]
    : [];

  return (
    <div className="mx-auto max-w-[1440px] px-margin-mobile py-lg md:px-margin-desktop">
      <div className="mb-lg">
        <h1 className="font-headline-xl font-bold text-cohere-primary">Dashboard 概览</h1>
        <p className="mt-1 font-body-main text-cohere-muted">实时系统状态与数据分析</p>
      </div>

      {/* Stat bento */}
      <div className="grid grid-cols-1 gap-gutter sm:grid-cols-2 lg:grid-cols-5">
        <StatCard label="总用户数" value={stats ? formatK(stats.totalUsers) : "—"} icon="group" />
        <StatCard label="总帖子数" value={stats ? stats.totalPosts.toLocaleString() : "—"} icon="article" />
        <StatCard label="AI 回复数" value={stats ? formatK(stats.aiReplies) : "—"} icon="forum" />
        <StatCard
          label="今日 AI 任务"
          value={stats ? stats.todayAiTasks.toLocaleString() : "—"}
          icon="memory"
          variant="secondary"
        />
        <StatCard
          label="失败任务"
          value={stats ? String(stats.failedTasks) : "—"}
          icon="error"
          variant="error"
        />
      </div>

      {/* Charts */}
      <div className="mt-xl grid grid-cols-1 gap-gutter lg:grid-cols-3">
        <div className="card-base p-lg lg:col-span-2">
          <h2 className="mb-md font-feature-title text-cohere-primary">7 日发帖趋势</h2>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={trend} margin={{ top: 8, right: 8, left: -16, bottom: 0 }}>
                <CartesianGrid stroke="#d9d9dd" strokeDasharray="3 3" vertical={false} />
                <XAxis
                  dataKey="label"
                  tick={{ fill: "#5f5e64", fontSize: 12, fontFamily: "'JetBrains Mono', monospace" }}
                  axisLine={{ stroke: "#d9d9dd" }}
                  tickLine={false}
                />
                <YAxis
                  tick={{ fill: "#5f5e64", fontSize: 12, fontFamily: "'JetBrains Mono', monospace" }}
                  axisLine={false}
                  tickLine={false}
                />
                <Tooltip
                  contentStyle={{
                    background: "#ffffff",
                    border: "1px solid #d9d9dd",
                    borderRadius: 8,
                    fontFamily: "'JetBrains Mono', monospace",
                    fontSize: 12,
                  }}
                />
                <Line
                  type="monotone"
                  dataKey="value"
                  stroke="#212121"
                  strokeWidth={2}
                  dot={{ r: 4, fill: "#fff", stroke: "#212121" }}
                  activeDot={{ r: 6 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div className="card-base p-lg">
          <h2 className="mb-md font-feature-title text-cohere-primary">AI 任务状态分布</h2>
          <div className="flex h-64 w-full flex-col justify-center gap-md">
            {statusRows.map((row) => (
              <div key={row.name}>
                <div className="mb-1 flex justify-between font-label-mono text-cohere-on-surface">
                  <span>{row.name}</span>
                  <span>{row.value}%</span>
                </div>
                <div className="h-2 rounded-full bg-cohere-surface-variant">
                  <div className="h-2 rounded-full bg-cohere-action-blue" style={{ width: `${row.value}%` }} />
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Service status */}
      <div className="card-base mt-xl p-lg">
        <h2 className="mb-md font-feature-title text-cohere-primary">系统服务状态</h2>
        <div className="grid grid-cols-2 gap-md md:grid-cols-4">
          {services.map((s) => (
            <div
              key={s.name}
              className="flex items-center gap-sm rounded-lg border border-cohere-hairline bg-cohere-success p-sm"
            >
              <span className="status-dot bg-cohere-secondary" />
              <span className="font-label-mono-bold text-cohere-on-secondary-container">{s.name}</span>
              <span className="ml-auto font-micro text-cohere-on-secondary-container">{s.metric}</span>
            </div>
          ))}
        </div>
      </div>

      {/* Asymmetric tables */}
      <div className="mt-xl grid grid-cols-1 gap-gutter lg:grid-cols-12">
        <div className="card-base flex flex-col overflow-hidden p-lg lg:col-span-7">
          <div className="mb-md flex items-center justify-between">
            <h2 className="font-feature-title text-cohere-primary">最新发布帖子</h2>
            <button type="button" className="btn-link">查看全部 →</button>
          </div>
          <table className="w-full text-left">
            <thead className="border-b border-cohere-hairline font-label-mono text-cohere-muted">
              <tr>
                <th className="py-1">标题</th>
                <th className="py-1">作者</th>
                <th className="py-1">时间</th>
                <th className="py-1 text-right">状态</th>
              </tr>
            </thead>
            <tbody className="font-body-main text-cohere-on-surface">
              {recentPosts.map((p) => (
                <tr key={p.id} className="border-b border-cohere-hairline transition-colors hover:bg-cohere-surface-low">
                  <td className="py-sm">{p.title}</td>
                  <td className="py-sm">{p.author}</td>
                  <td className="py-sm font-micro text-cohere-muted">{p.relativeTime}</td>
                  <td className="py-sm text-right">
                    <span
                      className={`rounded-full px-1 py-0.5 font-label-mono ${
                        p.status === "published"
                          ? "bg-cohere-success text-cohere-secondary"
                          : "bg-cohere-surface-variant text-cohere-muted"
                      }`}
                    >
                      {p.status === "published" ? "已发布" : "审核中"}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <div className="flex flex-col gap-gutter lg:col-span-5">
          <div className="card-base flex-1 p-lg">
            <h2 className="mb-md font-feature-title text-cohere-primary">最新 AI 任务</h2>
            <div className="flex flex-col gap-sm">
              {recentTasks.map((t) => (
                <div
                  key={t.id}
                  className="flex items-center justify-between rounded-lg border border-cohere-hairline p-sm transition-colors hover:bg-cohere-surface-low"
                >
                  <div className="flex items-center gap-sm">
                    <MaterialIcon
                      name={t.icon}
                      className={
                        t.status === "PROCESSING"
                          ? "text-cohere-action-blue"
                          : t.status === "COMPLETED"
                            ? "text-cohere-secondary"
                            : "text-cohere-error"
                      }
                    />
                    <span className="font-body-main text-cohere-on-surface">{t.label}</span>
                  </div>
                  <span
                    className={`rounded px-1 py-0.5 font-label-mono ${
                      t.status === "PROCESSING"
                        ? "bg-cohere-action-blue text-white"
                        : t.status === "COMPLETED"
                          ? "bg-cohere-success text-cohere-secondary"
                          : "bg-cohere-error-container text-cohere-error"
                    }`}
                  >
                    {t.status === "PROCESSING" ? "Running" : t.status === "COMPLETED" ? "Completed" : "Failed"}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <div className="flex-1 rounded-lg border border-cohere-hairline bg-cohere-surface-container p-lg">
            <h2 className="mb-md font-feature-title text-cohere-primary">AI 决策日志</h2>
            <div className="relative ml-2 space-y-md border-l border-dotted border-cohere-slate">
              {decisionTimeline.map((entry, idx) => (
                <div key={idx} className="relative pl-sm">
                  <span className="absolute -left-[5px] top-1.5 h-2 w-2 rounded-full bg-cohere-slate" />
                  <div className="font-label-mono text-micro text-cohere-muted">{entry.time}</div>
                  <p className="font-caption text-cohere-on-surface">{entry.message}</p>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function formatK(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}
