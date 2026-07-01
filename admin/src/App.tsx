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
import { mockDataProvider } from "./dataProvider/mockDataProvider";
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
    colorError: "#ba1a1a",
    colorWarning: "#ff7759",
    colorBgBase: "#fbf9f4",
    colorTextBase: "#1b1c19",
    colorBorder: "#d9d9dd",
    colorBgContainer: "#ffffff",
    colorBgLayout: "#fbf9f4",
    borderRadius: 8,
    fontFamily: "'Hanken Grotesk', Inter, ui-sans-serif, system-ui, sans-serif",
    fontFamilyCode: "'JetBrains Mono', ui-monospace, monospace",
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
              dataProvider={mockDataProvider}
              routerProvider={routerProvider}
              resources={[
                { name: "agents", list: "/agents" },
                { name: "tasks", list: "/tasks" },
                { name: "decisionLogs", list: "/decisions" },
                { name: "posts", list: "/posts" },
              ]}
              options={{ syncWithLocation: true }}
            >
              <Routes>
                <Route element={<AdminLayout />}>
                  <Route path="/" element={<DashboardPage />} />
                  <Route path="/users" element={<UsersManagePage />} />
                  <Route path="/posts" element={<PostsManagePage />} />
                  <Route path="/agents" element={<AIAgentsPage />} />
                  <Route path="/tasks" element={<AITasksPage />} />
                  <Route path="/decisions" element={<AIDecisionLogsPage />} />
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
