import { Refine } from "@refinedev/core";
import { RefineThemes } from "@refinedev/antd";
import routerProvider from "@refinedev/react-router-v6";
import { ConfigProvider, App as AntdApp } from "antd";
import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import AdminLayout from "./components/layout/AdminLayout";
import DashboardPage from "./pages/DashboardPage";
import AIAgentsPage from "./pages/AIAgentsPage";
import AITasksPage from "./pages/AITasksPage";
import AIDecisionLogsPage from "./pages/AIDecisionLogsPage";
import PostsManagePage from "./pages/PostsManagePage";
import UsersManagePage from "./pages/UsersManagePage";
import NotFoundPage from "./pages/NotFoundPage";
import LoginPage from "./pages/LoginPage";
import { CommentsPage, TagsPage, PreferencesPage } from "./pages/SimpleResourcePage";
import { dataProvider } from "./api/dataProvider";
import { authProvider } from "./api/authProvider";
import { accessControlProvider } from "./providers/accessControlProvider";
import "./styles/index.css";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { refetchOnWindowFocus: false, retry: 1, staleTime: 30_000 },
  },
});

/**
 * Ant Design theme overridden with the Cohere token system so Refine/AntD
 * components (Drawer, Table, Switch, Slider…) inherit the research-lab look.
 */
const cohereTheme = {
  ...RefineThemes.Blue,
  token: {
    colorPrimary: "#000000",
    colorInfo: "#1863dc",
    colorSuccess: "#35675d",
    colorError: "#8f1111",
    colorWarning: "#a83b24",
    colorBgBase: "#fbf9f4",
    colorTextBase: "#1b1c19",
    colorBorder: "#d9d9dd",
    colorBgContainer: "#ffffff",
    colorBgLayout: "#fbf9f4",
    borderRadius: 8,
    fontFamily: "'CohereText', 'Space Grotesk', 'Unica77', Inter, ui-sans-serif, system-ui, sans-serif",
    fontFamilyCode: "'CohereMono', 'JetBrains Mono', ui-monospace, monospace",
    fontSize: 14,
  },
  components: {
    Drawer: {
      colorBgElevated: "#fbf9f4",
      headerBg: "#f5f3ee",
    },
    Table: {
      headerBg: "#f0eee9",
      headerColor: "#1b1c19",
      borderColor: "#d9d9dd",
      rowHoverBg: "#f5f3ee",
    },
    Switch: {
      colorPrimary: "#000000",
      colorPrimaryHover: "#212121",
    },
  },
};

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <ConfigProvider theme={cohereTheme}>
          <AntdApp>
            <Refine
              dataProvider={dataProvider}
              authProvider={authProvider}
              accessControlProvider={accessControlProvider}
              routerProvider={routerProvider}
              resources={[
                { name: "users", list: "/users" },
                { name: "posts", list: "/posts" },
                { name: "comments", list: "/comments" },
                { name: "agents", list: "/agents", edit: "/agents/:id" },
                { name: "tasks", list: "/tasks", show: "/tasks/:id" },
                { name: "decisionLogs", list: "/decisions" },
                { name: "tags", list: "/tags" },
                { name: "preferences", list: "/preferences" },
              ]}
              options={{ syncWithLocation: true, disableTelemetry: true }}
            >
              <Routes>
                <Route path="/login" element={<LoginPage />} />
                <Route element={<AdminLayout />}>
                  <Route path="/" element={<DashboardPage />} />
                  <Route path="/users" element={<UsersManagePage />} />
                  <Route path="/posts" element={<PostsManagePage />} />
                  <Route path="/comments" element={<CommentsPage />} />
                  <Route path="/agents" element={<AIAgentsPage />} />
                  <Route path="/tasks" element={<AITasksPage />} />
                  <Route path="/decisions" element={<AIDecisionLogsPage />} />
                  <Route path="/tags" element={<TagsPage />} />
                  <Route path="/preferences" element={<PreferencesPage />} />
                  <Route path="/profile" element={<Navigate to="/" replace />} />
                  <Route path="*" element={<NotFoundPage />} />
                </Route>
              </Routes>
            </Refine>
          </AntdApp>
        </ConfigProvider>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
