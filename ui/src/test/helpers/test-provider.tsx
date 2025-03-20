import { CreateTableDialogProvider } from "@/context/create-table.tsx";
import { SidebarProvider } from "@/context/sidebar.tsx";
import { TablesProvider } from "@/context/tables.tsx";
import * as React from "react";

export function TestProvider({ children }: { children: React.ReactNode }) {
  return (
    <SidebarProvider>
      <TablesProvider>
        <CreateTableDialogProvider>{children}</CreateTableDialogProvider>
      </TablesProvider>
    </SidebarProvider>
  );
}
