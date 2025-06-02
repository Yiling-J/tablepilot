import { deleteTable, TableCreateRequest } from "@/actions";
import { ImportFileDialog } from "@/components/dialog/import-file";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
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
import { FileIcon, PlusIcon } from "@radix-ui/react-icons";
import { Trash2 } from "lucide-react";
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
  const { openNewTableDialog, withForm, withRows } = useCreateTableDialog();
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
              <Card
                key={table.id}
                className="h-60 flex flex-col rounded-lg bg-background border border-gray-400/30 hover:bg-muted-foreground/5 cursor-pointer"
              >
                <div onClick={() => navigate(`/tables/${table.id}`)} className="flex flex-col flex-grow p-4">
                  <CardHeader className="p-0 pb-2"> {/* Adjusted padding */}
                    <div className="text-lg font-semibold truncate">{table.name}</div> {/* Adjusted style */}
                  </CardHeader>
                  <CardContent className="grow mt-2 p-0">
                    <p className="line-clamp-4">{table.description}</p>
                  </CardContent>
                </div>
                <CardFooter className="px-4 py-3 border-t border-gray-400/30"> {/* Adjusted padding */}
                  <AlertDialog>
                    <AlertDialogTrigger asChild>
                      <Button
                        variant="ghost"
                        size="icon"
                        title="Delete Table"
                        className="text-destructive hover:text-destructive hover:bg-destructive/10 ml-auto"
                        onClick={(e) => e.stopPropagation()} // Prevent navigation
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </AlertDialogTrigger>
                    <AlertDialogContent onClick={(e) => e.stopPropagation()}>
                      <AlertDialogHeader>
                        <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                        <AlertDialogDescription>
                          This action cannot be undone. This will permanently delete the table.
                        </AlertDialogDescription>
                      </AlertDialogHeader>
                      <AlertDialogFooter>
                        <AlertDialogCancel>Cancel</AlertDialogCancel>
                        <AlertDialogAction
                          onClick={async () => {
                            await deleteTable(table.id);
                            await fetchTables(); // Refetch after delete
                            refreshTables(); // Context refresh
                          }}
                        >
                          Delete
                        </AlertDialogAction>
                      </AlertDialogFooter>
                    </AlertDialogContent>
                  </AlertDialog>
                </CardFooter>
              </Card>
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
