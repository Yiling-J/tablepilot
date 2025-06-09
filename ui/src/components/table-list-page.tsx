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
import { TablepilotHeader } from "./header.tsx";
import { ScrollArea } from "./ui/scroll-area.tsx";

export function TableListPage() {
  return (
    <div className="grow overflow-auto h-full flex flex-col">
      <ModeToggle hide={true} />
      <TablepilotHeader title="Tablepilot" currentTab="tables" />
      <ScrollArea className="h-[calc(100vh-120px)]">
        <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8 py-12">
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
    withRows,
    withTable,
    withSubmitCallback,
  } = useCreateTableDialog();
  const [importCSVOpen, setImportCSVOpen] = useState(false);
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
  }, []);

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
    <div className="grow overflow-auto h-full flex flex-col">
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
        <Card className="flex flex-col cursor-pointer h-60 min-w-72 border-dashed overflow-hidden">
          <div
            className="flex flex-col items-center justify-center hover:bg-muted-foreground/5 transition-all w-full h-full flex-1 hover:h-[70%] peer"
            onClick={() => {
              openNewTableDialog();
            }}
          >
            <PlusIcon className="w-5 h-5 mr-2 mb-2" />
            <span>Add New Table</span>
          </div>

          <div
            className="flex flex-col items-center justify-center hover:bg-muted-foreground/5 transition-all w-full h-[30%] peer-hover:h-[30%] hover:h-[70%] border-t rounded-t-xl"
            onClick={() => setImportCSVOpen(true)}
          >
            <div className="flex items-center">
              <FileIcon className="w-4 h-4 mr-2" />
              <span>Import</span>
            </div>
            <p className="text-xs pt-2 text-gray-500">
              formats: csv, png, jpg, jpeg
            </p>
          </div>
        </Card>
      </div>
    </div>
  );
}
