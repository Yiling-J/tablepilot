import {
    deleteTable,
    deleteWorkflow,
    getTables,
    getWorkflow,
    getWorkflows,
    TableCreateRequest,
    TableInfo,
    Workflow,
    WorkflowInfo,
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
import { cn } from "@/lib/utils";
import { FileIcon, PlusIcon } from "@radix-ui/react-icons";
import { SettingsIcon } from "lucide-react";
import { useEffect, useState, useCallback } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { ModeToggle } from "./darkmode";
import WorkflowBuilderDialog from "./dialog/workflow/builder.tsx";
import WorkflowExecutionDialog from "./dialog/workflow/workflow.tsx";
import { TablepilotHeader } from "./header.tsx";

export function TableListPage() {
  const location = useLocation();
  const navigate = useNavigate();

  // Initial tab state based on current URL path
  const [tab, setTab] = useState(() =>
    location.pathname.startsWith("/workflows") ? "workflows" : "tables"
  );

  // Effect to update tab state if URL changes (e.g., browser back/forward)
  useEffect(() => {
    if (location.pathname.startsWith("/workflows")) {
      setTab("workflows");
    } else {
      // Defaults to "tables" for "/tables" or any other path reaching here
      setTab("tables");
    }
  }, [location.pathname]);

  const handleTabChange = (newTab: string) => {
    navigate(`/${newTab}`); // newTab will be "tables" or "workflows"
  };

  return (
    <div className="grow overflow-auto h-full flex flex-col">
      <ModeToggle hide={true} />
      <TablepilotHeader
        title="Tablepilot"
        currentTab={tab}
        onTabChange={handleTabChange}
        // onRefresh={handleRefresh} // Removed
      />
      <div className="max-w-6xl mx-auto px-4 py-8 sm:px-6 lg:px-8 py-12">
        {/* refreshKey prop removed from TableList and WorkflowList */}
        {/* Wrapped conditional rendering for fade-in animation */}
        <div key={tab} className="tab-content-container">
          {tab === "tables" ? <TableList /> : <WorkflowList />}
        </div>
      </div>
    </div>
  );
}

// interface TableListProps { // Removed refreshKey
//   refreshKey: number;
// }

function TableList(/*{ refreshKey }: TableListProps*/) { // Removed refreshKey from props destructuring
  const [tables, setTables] = useState<TableInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const { openNewTableDialog, withForm, withRows } = useCreateTableDialog();
  const [importCSVOpen, setImportCSVOpen] = useState(false);
  const { refreshTables } = useTables();
  const navigate = useNavigate(); // This instance is for TableList's own navigation needs, separate from TableListPage

  const fetchTables = useCallback(async () => {
    setLoading(true);
    try {
      const response = await getTables();
      setTables(response.tables ?? []);
    } catch (error) {
      console.error("Failed to fetch tables:", error);
      setTables([]); // Set to empty array on error
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTables();
    // refreshTables() from context was removed as fetchTables handles local data.
    // If global context needs refresh, it should be handled more explicitly if needed,
    // or the component consuming global context should use refreshTables itself.
  }, [fetchTables]); // refreshKey removed from dependency array

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
                <div className="text-xl font-bold truncate">{table.name}</div>
                <div className="grow mt-2">
                  <p className="line-clamp-4">{table.description}</p>
                </div>

                <div className="self-end">
                  <Button
                    variant="destructive"
                    onClick={async (e) => {
                      e.stopPropagation();
                      // No need to setLoading(true) here as fetchTables handles it.
                      await deleteTable(table.id);
                      await fetchTables(); // Refetch after delete
                      refreshTables(); // Context refresh
                      // No need to setLoading(false) here
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
  );
}

// interface WorkflowListProps { // Removed refreshKey
//   refreshKey: number;
// }

function WorkflowList(/*{ refreshKey }: WorkflowListProps*/) { // Removed refreshKey from props destructuring
  const [workflows, setWorkflows] = useState<WorkflowInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [workflow, setWorkflow] = useState<undefined | Workflow>(undefined);
  const [runWorkflowOpen, setRunWorkflowOpen] = useState(false);
  const [WorkflowBuilderOpen, setRunWorkflowBuilderOpen] = useState(false);

  const refreshWorkflows = useCallback(async () => {
    setLoading(true);
    try {
      const wf = await getWorkflows();
      setWorkflows(wf.workflows ?? []); // Ensure workflows is an array
    } catch (error) {
      console.error("Failed to fetch workflows:", error);
      setWorkflows([]); // Set to empty array on error
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refreshWorkflows();
  }, [refreshWorkflows]); // refreshKey removed from dependency array

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <WorkflowExecutionDialog
        workflow={workflow}
        open={runWorkflowOpen}
        onOpenChange={setRunWorkflowOpen}
      />
      <WorkflowBuilderDialog
        id={workflow?.id}
        workflow={workflow}
        open={WorkflowBuilderOpen}
        onOpenChange={setRunWorkflowBuilderOpen}
        onSave={refreshWorkflows}
      />
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
        : workflows.map((wf) => (
            <div
              key={wf.id}
              className="h-60 flex flex-col rounded-lg bg-background p-4 border border-gray-400/30 hover:bg-muted-foreground/5 cursor-pointer"
              onClick={async () => {
                const w = await getWorkflow(wf.id);
                setWorkflow(w);
                setRunWorkflowOpen(true);
              }}
            >
              <div className="text-xl font-bold truncate">{wf.name}</div>
              <div className="grow mt-2">
                <p className="line-clamp-4">{wf.description}</p>
              </div>

              <div className="flex justify-between">
                <Button
                  variant="outline"
                  size="icon"
                  onClick={async (e) => {
                    e.stopPropagation();
                    const w = await getWorkflow(wf.id);
                    setWorkflow(w);
                    setRunWorkflowBuilderOpen(true);
                  }}
                >
                  <SettingsIcon />
                </Button>
                <Button
                  variant="destructive"
                  onClick={async (e) => {
                    e.stopPropagation();
                    // No need to setLoading(true) here as refreshWorkflows handles it.
                    await deleteWorkflow(wf.id);
                    await refreshWorkflows(); // Refetch after delete
                    // No need to setLoading(false) here
                  }}
                >
                  Delete
                </Button>
              </div>
            </div>
          ))}
      <Card className="flex flex-col cursor-pointer h-60 min-w-72 border-dashed overflow-hidden">
        <div
          className="flex flex-col items-center justify-center hover:bg-muted-foreground/5 transition-all w-full h-full flex-1"
          onClick={() => {
            setRunWorkflowBuilderOpen(true);
          }}
        >
          <PlusIcon className="w-5 h-5 mr-2 mb-2" />
          <span>Add New Workflow</span>
        </div>
      </Card>
    </div>
  );
}
