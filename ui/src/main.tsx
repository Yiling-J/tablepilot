import "@material-symbols/font-400/rounded.css";
import React from "react";
import ReactDOM from "react-dom/client";
// ModelManager import removed as it's now used within ModelManagerPageWrapper
import { ModelManagerPageWrapper } from "./components/models/model-manager-page-wrapper.tsx"; // Added
import { TableListPage } from "./components/table-list-page.tsx";
import { TablePage } from "./components/table.tsx";
import { WorkflowListPage } from "./components/workflow-list-page.tsx";
import { CreateTableDialogProvider } from "./context/create-table.tsx";
import "./index.css";

import { Toaster as ShadcnToaster } from "@/components/ui/toaster";
import { ErrorBoundary } from "react-error-boundary";
import { Toaster } from "react-hot-toast";

import {
    Navigate,
    Outlet,
    RouterProvider,
    createBrowserRouter,
} from "react-router-dom";
import { Sidebar } from "./components/sidebar.tsx";
import { SidebarProvider } from "./context/sidebar.tsx";
import { TablesProvider } from "./context/tables.tsx";

const router = createBrowserRouter([
  {
    element: (
      <ErrorBoundary fallback={<div>Something went wrong</div>}>
        <SidebarProvider>
          <TablesProvider>
            <CreateTableDialogProvider>
              <Toaster />
              <ShadcnToaster />
              <Outlet />
            </CreateTableDialogProvider>
          </TablesProvider>
        </SidebarProvider>
      </ErrorBoundary>
    ),
    children: [
      {
        element: (
          <div className="bg-muted/50 scrollbar-thumb-rounded-full scrollbar-track-rounded-full scrollbar scrollbar-thumb-stone-500 scrollbar-track-background">
            <div className="flex h-screen w-screen">
              <Sidebar className="flex" />
              <Outlet />
            </div>
          </div>
        ),
        children: [
          // Routes that WILL have the Sidebar
          { path: "/tables/:id", element: <TablePage /> },
          // { path: "/models", element: <ModelManager searchTerm="" /> }, // Route moved
          // Add other routes that need the sidebar here
        ],
      },
      // Routes that will NOT have the Sidebar (rendered directly into the root Outlet)
      { path: "/tables", element: <TableListPage /> },
      { path: "/workflows", element: <WorkflowListPage /> },
      { path: "/models", element: <ModelManagerPageWrapper /> }, // Route added here
      {
        path: "/", // Root redirect
        element: <Navigate to="/tables" replace />,
      },
    ],
  },
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>,
);
