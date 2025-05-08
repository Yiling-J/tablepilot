import {
    Cell,
    ColumnDef,
    RowData,
    flexRender,
    getCoreRowModel,
    useReactTable,
} from "@tanstack/react-table";

import { CellTextDialog } from "@/components/dialog/cell-text.tsx";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "@/components/ui/table";
import { JSONObject } from "@/json";
import { cn } from "@/lib/utils";
import { SizeIcon } from "@radix-ui/react-icons";
import { ScrollArea } from "@radix-ui/react-scroll-area";
import { MutableRefObject, memo, useRef, useState } from "react";
import { Checkbox } from "../ui/checkbox";

/* eslint-disable @typescript-eslint/no-unused-vars */
declare module "@tanstack/react-table" {
  interface ColumnMeta<TData extends RowData, TValue> {
    columnType: string;
  }
}

function TableCellEx({
  cell,
  setExpandCellOpen,
  expandCellTextRef,
  selected,
  onRowSelectChange,
}: {
  cell: Cell<JSONObject, string>;
  setExpandCellOpen: (v: boolean) => void;
  expandCellTextRef: MutableRefObject<string>;
  onRowSelectChange: (row: string, selected: boolean) => void;
  selected: boolean;
}) {
  const [hoverCell, setHoverCell] = useState(false);
  const columnType = cell.column.columnDef.meta?.columnType;

  const hover = useRef(() => {
    setHoverCell(true);
  });
  const unHover = useRef(() => {
    setHoverCell(false);
  });
  const expandRef = useRef(() => {
    expandCellTextRef.current = cell.renderValue() ?? "";
    setExpandCellOpen(true);
  });

  const showExpand = useRef(
    ["string", "array"].includes(columnType!.toString()),
  );

  if (
    (selected && cell.column.id == "rowIndex") ||
    (hoverCell && cell.column.id == "rowIndex")
  ) {
    return (
      <TableCell
        key={cell.id}
        className="border-l border-b last:border-r border-sky-900 max-w-lg max-h-80 whitespace-pre-wrap relative z-0"
        onMouseEnter={hover.current}
        onMouseLeave={unHover.current}
      >
        <div className="w-10 h-4">
          <Checkbox
            id="row-selected"
            checked={selected}
            onCheckedChange={(v) => {
              onRowSelectChange(cell.row.id, v as boolean);
            }}
          />
        </div>
      </TableCell>
    );
  }

  return (
    <TableCell
      key={cell.id}
      className={cn(
        "border-l border-b last:border-r border-sky-900 max-w-lg max-h-80 whitespace-pre-wrap relative z-0",
        cell.column.id == "rowIndex" ? "cursor-pointer" : "",
      )}
      onMouseEnter={hover.current}
      onMouseLeave={unHover.current}
    >
      {flexRender(cell.column.columnDef.cell, cell.getContext())}
      {hoverCell && (
        <div className="absolute bottom-0 right-0 flex pr-1 pb-1">
          {showExpand.current && (
            <Button
              size="icon"
              className="rounded-full border hover:scale-100 transition-transform duration-50 scale-90 group hover:bg-secondary"
              variant="secondary"
              onClick={expandRef.current}
            >
              <SizeIcon className="group-hover:scale-150 transition-transform duration-50" />
            </Button>
          )}
        </div>
      )}
    </TableCell>
  );
}

interface DataGridProps {
  columns: ColumnDef<JSONObject, string>[];
  data: JSONObject[];
  selectedRows: string[];
  onRowSelectChange: (row: string, selected: boolean) => void;
  loading: string[];
}

const TableCellExMemo = memo(TableCellEx);

export function DataGrid({
  columns,
  data,
  selectedRows,
  onRowSelectChange,
  loading,
}: DataGridProps) {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getRowId: (row) => row.__id__ as string,
  });
  const [expandCellOpen, setExpandCellOpen] = useState(false);
  const expandCellTextRef = useRef("");

  return (
    <ScrollArea>
      <CellTextDialog
        text={expandCellTextRef.current}
        isOpen={expandCellOpen}
        setIsOpen={setExpandCellOpen}
      />
      <Table className="border-separate border-spacing-0 mb-10">
        <TableHeader className="sticky top-0 bg-secondary z-10">
          {table.getHeaderGroups().map((headerGroup) => (
            <TableRow key={headerGroup.id}>
              {headerGroup.headers.map((header) => {
                return (
                  <TableHead
                    key={header.id}
                    className="border-l border-b border-t last:border-r border-sky-900 first:min-w-6 pl-0 pr-4 min-w-32 min-h-6 max-h-32 cursor-pointer hover:bg-zinc-200 dark:hover:bg-zinc-600"
                  >
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext(),
                        )}
                  </TableHead>
                );
              })}
            </TableRow>
          ))}
        </TableHeader>
        <TableBody>
          {table.getRowModel().rows.map((row) => (
            <TableRow
              key={row.id}
              data-state={row.getIsSelected() && "selected"}
            >
              {loading.findIndex((v) => v === row.id) > -1 && (
                <TableCell
                  colSpan={columns.length}
                  className="border-l border-b last:border-r border-sky-900 max-w-lg whitespace-pre-wrap relative z-0"
                >
                  <Skeleton className="h-16" />
                </TableCell>
              )}
              {loading.findIndex((v) => v === row.id) < 0 &&
                row
                  .getVisibleCells()
                  .map((cell) => (
                    <TableCellExMemo
                      cell={cell}
                      key={cell.id}
                      setExpandCellOpen={setExpandCellOpen}
                      expandCellTextRef={expandCellTextRef}
                      selected={
                        selectedRows.findIndex((v) => v === cell.row.id) > -1
                      }
                      onRowSelectChange={onRowSelectChange}
                    />
                  ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </ScrollArea>
  );
}
