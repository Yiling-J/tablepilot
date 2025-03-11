import * as React from "react";

import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import {
    ChevronLeftIcon,
    Cross1Icon,
    GridIcon,
    PlusIcon,
    ReloadIcon,
} from "@radix-ui/react-icons";

import { deleteTable } from "@/actions";
import { ModeToggle } from "@/components/darkmode";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { useCreateTableDialog } from "@/context/create-table";
import { useSidebar } from "@/context/sidebar";
import { useTables } from "@/context/tables";
import { useEffect, useRef, useState } from "react";
import { Sidebar as ReactSidebar } from "react-pro-sidebar";
import { useNavigate, useParams } from "react-router-dom";

interface delInfo {
  name: string;
  tableId: string;
  viewId: string;
}

export type SidebarProps = React.ComponentProps<"div">;

export function Sidebar({ className }: SidebarProps) {
  const navigate = useNavigate();
  const { openNewTableDialog } = useCreateTableDialog();
  const [hoverTable, setHoverTable] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const viewTableRef = useRef("");
  const delRef = useRef<delInfo>({ tableId: "", viewId: "", name: "" });
  const { id, viewid } = useParams();
  const { tables, refreshTables } = useTables();
  const { collapsed, setCollapsed } = useSidebar();

  useEffect(() => {
    refreshTables();
  }, []);

  if (collapsed) {
    return (
      <div className="fixed left-0 top-0 p-7">
        <ModeToggle hide={true} />
      </div>
    );
  }

  const handleDeleteTable = async () => {
    setIsLoading(true);
    if (
      delRef.current.tableId.length > 0 &&
      delRef.current.viewId.length == 0
    ) {
      await deleteTable(delRef.current.tableId);
    }

    await refreshTables();
    setDialogOpen(false);
    setIsLoading(false);

    const currentPage = `${id ?? ""}>${viewid ?? ""}`;
    const delPage = `${delRef.current.tableId}>${delRef.current.viewId}`;

    if (currentPage == delPage) {
      navigate("/");
    }
  };

  const tableNodes = tables.map((t) => (
    <div className="flex flex-col" key={t.id}>
      <li
        key={t.id}
        className={cn(
          "cursor-pointer py-2 truncate text-sm flex items-center bg-primary/0 hover:bg-primary/10 pl-7 h-12",
          `${id}${viewid ?? ""}` == t.id ? "bg-primary/15" : "",
        )}
        onClick={() => {
          navigate(`/tables/${t.id}`);
        }}
        onMouseEnter={() => {
          setHoverTable(t.id);
          viewTableRef.current = t.id;
        }}
        onMouseLeave={() => setHoverTable("")}
      >
        <div className="flex flex-row justify-between w-full pr-7 items-center">
          <div className="flex items-center">
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={(e) => {
                e.stopPropagation();
              }}
            >
              <GridIcon />
            </Button>
            <p className="ml-2">{t.name}</p>
          </div>

          <div
            className={cn(
              "flex items-center",
              t.id == hoverTable ? "" : "invisible",
            )}
          >
            <Button
              variant="ghost"
              size="icon"
              className="w-7 h-7"
              onClick={(e) => {
                e.stopPropagation();
                delRef.current.tableId = t.id;
                delRef.current.viewId = "";
                delRef.current.name = t.name;
                setDialogOpen(true);
              }}
            >
              <Cross1Icon />
            </Button>
          </div>
        </div>
      </li>
    </div>
  ));

  return (
    <ReactSidebar
      width="270px"
      breakPoint="md"
      collapsed={collapsed}
      toggled={true}
      collapsedWidth="0px"
      rootStyles={{
        borderColor: "rgba(12, 12, 12, 0.1)",
      }}
    >
      <Dialog
        open={dialogOpen}
        onOpenChange={(open) => {
          setDialogOpen(open);
          setTimeout(() => {
            if (!open) {
              document.body.style.pointerEvents = "";
            }
          }, 200);
        }}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirm Delete {delRef.current.name}</DialogTitle>
          </DialogHeader>
          <DialogDescription>
            Are you sure you want to delete this{" "}
            {delRef.current.viewId.length == 0 ? "table" : "view"}?
          </DialogDescription>
          <DialogFooter>
            <Button
              onClick={() => {
                setDialogOpen(false);
                setTimeout(() => {
                  document.body.style.pointerEvents = "";
                }, 200);
              }}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              onClick={async () => {
                await handleDeleteTable();
                setTimeout(() => {
                  document.body.style.pointerEvents = "";
                }, 200);
              }}
              disabled={isLoading}
            >
              {isLoading ? <ReloadIcon className="animate-spin" /> : "Confirm"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <div
        className={cn(
          className,
          "h-full flex-col bg-gray-100 dark:bg-zinc-800",
        )}
      >
        <h1
          onClick={() => navigate("/")}
          className="scroll-m-20 text-xl font-extrabold tracking-widest pt-6 pl-5 cursor-pointer"
        >
          Tablepilot
        </h1>
        <p className="pl-5 pt-1 text-xs text-muted-foreground">
          Open-Source AI Table Generator
        </p>
        <Separator className="my-4 bg-gray-400/30 self-center w-[260px]" />
        <div
          className="mx-5 my-3 py-2 flex flex-row cursor-pointer font-semibold hover:bg-neutral-950/15 rounded-xl items-center"
          onClick={() => {
            openNewTableDialog();
          }}
        >
          <PlusIcon className="mr-3 ml-1" />
          Create New Table
        </div>

        <div>
          <div className="flex justify-between px-7 pt-4 pb-2 items-center">
            <p className="font-semibold text-base">Tables</p>{" "}
          </div>
        </div>
        <div className="grow pt-3 overflow-auto scrollbar">
          {tables.length > 0 && <ul className="">{tableNodes}</ul>}
        </div>
        <div className="pb-5 flex justify-between items-center">
          <ModeToggle hide={false} />
          <div className="flex">
            <button
              className="p-2  bg-transparent cursor-pointer h-7 w-7"
              onClick={() => {
                setCollapsed(true);
              }}
            >
              <ChevronLeftIcon />
            </button>
          </div>
        </div>
      </div>
    </ReactSidebar>
  );
}
