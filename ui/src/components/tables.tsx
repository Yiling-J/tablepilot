import {
    TableCreateRequest,
    TableInfo,
    deleteTable,
    getTables,
} from "@/actions";
import { ImportFileDialog } from "@/components/dialog/import-file";
import { Button } from "@/components/ui/button";
import {
    Card,
    CardContent,
    CardFooter,
    CardHeader,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { useCreateTableDialog } from "@/context/create-table";
import { useTables } from "@/context/tables";
import { JSONObject } from "@/json.ts";
import { FileIcon, PlusIcon, ReloadIcon } from "@radix-ui/react-icons";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { ModeToggle } from "./darkmode";
import { TablepilotHeader } from "./header.tsx";

export function TableListPage() {
  const [tables, setTables] = useState<TableInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const { openNewTableDialog, withForm, withRows } = useCreateTableDialog();
  const [importCSVOpen, setImportCSVOpen] = useState(false);
  const { refreshTables } = useTables();
  const navigate = useNavigate();

  useEffect(() => {
    setLoading(true);
    fetchTables().finally(() => setLoading(false));
  }, []);

  const fetchTables = async () => {
    const response = await getTables();
    setTables(response.tables ?? []);
  };

  return (
    <div className="grow overflow-auto h-full flex flex-col">
      <ImportFileDialog
        isOpen={importCSVOpen}
        setIsOpen={setImportCSVOpen}
        onNext={(form: TableCreateRequest, rows: JSONObject[]) => {
          withForm(form);
          withRows(rows);
          setImportCSVOpen(false);
          openNewTableDialog();
        }}
      />
      <ModeToggle hide={true} />
      <TablepilotHeader title="Tablepilot" />
      <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8 py-12">
        <div className="flex items-center justify-between mb-6">
          <h1 className="text-3xl font-bold">Tables</h1>
          <Button variant="outline" onClick={fetchTables}>
            <ReloadIcon className="w-6 h-6 mr-2" />
            Refresh
          </Button>
        </div>
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
                <div
                  key={table.id}
                  className="h-60 flex flex-col rounded-lg bg-background p-4 border border-gray-400/30 hover:bg-muted-foreground/5 cursor-pointer"
                  onClick={() => navigate(`/tables/${table.id}`)}
                >
                  <div className="text-xl font-bold">{table.name}</div>
                  <div className="grow mt-2">
                    <p className="line-clamp-4">{table.description}</p>
                  </div>

                  <div className="self-end">
                    <Button
                      variant="destructive"
                      onClick={async (e) => {
                        e.stopPropagation();
                        setLoading(true);
                        await deleteTable(table.id);
                        await fetchTables();
                        refreshTables();
                        setLoading(false);
                      }}
                    >
                      Delete
                    </Button>
                  </div>
                </div>
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
    </div>
  );
}
