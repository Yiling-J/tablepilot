import "@material-symbols/font-400/rounded.css";
import React from "react";
import ReactDOM from "react-dom/client";
import { TablePage } from "./components/table.tsx";
import { TableListPage } from "./components/tables.tsx";
import { CreateTableDialogProvider } from "./context/create-table.tsx";
import "./index.css";

import { ErrorBoundary } from "react-error-boundary";
import { Toaster } from "react-hot-toast";
import { Outlet, RouterProvider, createBrowserRouter } from "react-router-dom";
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
        children: [{ path: "/tables/:id", element: <TablePage /> }],
      },
      {
        path: "/",
        element: <TableListPage />,
      },
    ],
  },
]);

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>,
);
