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
import { useTables } from "@/context/tables";
import { JSONObject } from "@/json.ts";
import { FileIcon, PlusIcon } from "@radix-ui/react-icons";
import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ModeToggle } from "./darkmode";
import { Button } from "./ui/button.tsx";
import { TablepilotHeader } from "./header.tsx";
import { ScrollArea } from "./ui/scroll-area.tsx";

export function TableListPage() {
  const {
    openNewTableDialog,
    withForm,
    withRows,
    // withTable, // Removed as unused
    // withSubmitCallback, // Removed as unused
  } = useCreateTableDialog();
  const [importCSVOpen, setImportCSVOpen] = useState(false); // This is used by ImportFileDialog
  const { tables, refreshTables } = useTables();

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
    <div className="grow h-full flex flex-col">
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
      <div className="bg-background sticky top-0 z-10 pt-4 pb-1"> {/* Changed pb-2 to pb-1 and removed border-b */}
        <div className="max-w-6xl mx-auto px-4 sm:px-6 lg:px-8 flex justify-end space-x-2">
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
      <ScrollArea className="flex-grow">
        <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8">
          <div className="tab-content-container">
            <TableList />
          </div>
        </div>
      </ScrollArea>
    </div>
  );
}

function TableList() {
  const [loading, setLoading] = useState(true);
  const {
    openNewTableDialog,
    withForm,
    // withRows, // Removed as unused in TableList's instance
    withTable,
    withSubmitCallback,
  } = useCreateTableDialog();
  // const [importCSVOpen, setImportCSVOpen] = useState(false); // Removed as unused in TableList
  const { tables, refreshTables } = useTables();
  const navigate = useNavigate();

  const fetchTables = useCallback(async () => {
    setLoading(true);
    try {
      await refreshTables();
    } catch (error) {
      console.error("Failed to fetch tables:", error);
    } finally {
      setLoading(false);
    }
  }, [refreshTables]); // Added refreshTables to dependency array

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
      {/* Removed ImportFileDialog from here */}
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
          : tables.map((table) => (
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
        {/* The Add New Table and Import buttons/card has been removed from here */}
      </div>
    </div>
  );
}
