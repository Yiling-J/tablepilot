import { deleteTable, getTableSchema, TableCreateRequest } from "@/actions";
import { ImportFileDialog } from "@/components/dialog/import-file";
import {
    Card,
    CardContent,
    CardFooter,
    CardHeader,
} from "@/components/ui/card";
import { CommonCard } from "@/components/ui/common-card";
import { Skeleton } from "@/components/ui/skeleton";
import { useCreateTableDialog } from "@/context/create-table";
import { useDelayedLoading } from "@/hooks/use-delayed-loading";
import { useTables } from "@/context/tables";
import { JSONObject } from "@/json.ts";
import { FileIcon, PlusIcon } from "@radix-ui/react-icons";
import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ModeToggle } from "./darkmode";
import { TablepilotHeader } from "./header.tsx";
import { Button } from "./ui/button.tsx";
import { Input } from "./ui/input.tsx";
import { ScrollArea } from "./ui/scroll-area.tsx";

export function TableListPage() {
  const { openNewTableDialog, withForm, withRows } = useCreateTableDialog();
  const [importCSVOpen, setImportCSVOpen] = useState(false);
  const { tables, refreshTables } = useTables();
  const [searchQuery, setSearchQuery] = useState("");

  const fetchTables = useCallback(async () => {
    // No setLoading as TableList will handle its own loading state
    try {
      await refreshTables();
    } catch (error) {
      console.error("Failed to fetch tables:", error);
    }
  }, [refreshTables]);

  useEffect(() => {
    fetchTables();
  }, [fetchTables]);

  return (
    <div className="grow h-screen flex flex-col">
      <ModeToggle hide={true} />
      <TablepilotHeader title="Tablepilot" currentTab="tables" />
      <ImportFileDialog
        isOpen={importCSVOpen}
        setIsOpen={setImportCSVOpen}
        tables={tables}
        onNext={(form: TableCreateRequest, rows: JSONObject[]) => {
          withForm(form);
          withRows(rows);
          setImportCSVOpen(false);
          openNewTableDialog();
        }}
      />
      <div className="bg-background sticky top-0 z-10 pt-4 pb-1">
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 flex justify-between items-center space-x-4">
          <Input
            type="text"
            placeholder="Search tables..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="max-w-sm h-9 rounded-full"
          />
          <div className="flex space-x-2">
            <Button
              variant="outline"
              onClick={() => {
                openNewTableDialog();
              }}
            >
              <PlusIcon className="w-4 h-4 mr-2" />
              Add New Table
            </Button>
            <Button variant="outline" onClick={() => setImportCSVOpen(true)}>
              <FileIcon className="w-4 h-4 mr-2" />
              Import
            </Button>
          </div>
        </div>
      </div>
      <ScrollArea className="flex-grow">
        <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
          <div className="tab-content-container">
            <TableList searchQuery={searchQuery} />
          </div>
        </div>
      </ScrollArea>
    </div>
  );
}

interface TableListProps {
  searchQuery: string;
}

function TableList({ searchQuery }: TableListProps) {
  const [actualLoading, setActualLoading] = useState(true);
  const loading = useDelayedLoading(actualLoading, 500); // Using 500ms delay
  const { openNewTableDialog, withForm, withTable, withSubmitCallback } =
    useCreateTableDialog();
  const { tables, refreshTables } = useTables();
  const navigate = useNavigate();

  const fetchTables = useCallback(async () => {
    setActualLoading(true);
    try {
      await refreshTables();
    } catch (error) {
      console.error("Failed to fetch tables:", error);
    } finally {
      setActualLoading(false);
    }
  }, [refreshTables]);

  useEffect(() => {
    fetchTables();
  }, [fetchTables]);

  const handleEditTableClick = async (tableId: string) => {
    try {
      const schema = await getTableSchema(tableId);
      withForm(schema);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      withTable(tableId as any); // TODO: fix this type error
      withSubmitCallback(fetchTables);
      openNewTableDialog();
    } catch (error) {
      console.error("Failed to prepare table for editing:", error);
      // Optionally, show a user-facing error message here
    }
  };

  return (
    <div className="grow overflow-auto h-full flex flex-col pt-6">
      <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {loading
          ? Array.from({ length: 4 }).map((_, index) => (
              <Card key={index} className="w-80">
                <CardHeader>
                  <Skeleton className="h-6 w-3/4 mb-2" />
                </CardHeader>
                <CardContent>
                  <Skeleton className="h-4 w-full mb-2" />
                  <Skeleton className="h-4 w-full" />
                </CardContent>
                <CardFooter>
                  <Skeleton className="h-4 w-1/4" />
                </CardFooter>
              </Card>
            ))
          : tables
              .filter((table) =>
                table.name.toLowerCase().includes(searchQuery.toLowerCase()),
              )
              .map((table) => (
                <CommonCard
                  key={table.id}
                  name={table.name}
                  onClick={() => navigate(`/tables/${table.id}`)}
                  onEdit={() => handleEditTableClick(table.id)}
                  onDelete={async () => {
                    await deleteTable(table.id);
                    await fetchTables();
                    refreshTables();
                  }}
                >
                  <p className="line-clamp-4">{table.description}</p>
                </CommonCard>
              ))}
      </div>
    </div>
  );
}
