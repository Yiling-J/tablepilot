import { TableInfo, getTables } from "@/actions";
import { ReactNode, createContext, useContext, useState } from "react";

interface TablesContextValue {
  tables: TableInfo[];
  refreshTables: () => Promise<void>;
}

const TablesContext = createContext<TablesContextValue | undefined>(undefined);

export function useTables() {
  const context = useContext(TablesContext);
  if (!context) {
    throw new Error(
      "useCreateTableDialog must be used within a CreateTableDialogProvider",
    );
  }
  return context;
}

interface TablesProviderProps {
  children: ReactNode;
}

export function TablesProvider({ children }: TablesProviderProps) {
  const [tables, setTables] = useState<TableInfo[]>([]);

  const refreshTables = async () => {
    const response = await getTables();
    setTables(response.tables);
  };

  return (
    <TablesContext.Provider value={{ tables, refreshTables }}>
      {children}
    </TablesContext.Provider>
  );
}
