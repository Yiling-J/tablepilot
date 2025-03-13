import {
    Cell,
    ColumnDef,
    flexRender,
    getCoreRowModel,
    useReactTable,
} from "@tanstack/react-table";

import { CellTextDialog } from "@/components/dialog/cell-text.tsx";
import { Button } from "@/components/ui/button";
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

declare module "@tanstack/react-table" {
  interface ColumnMeta {
    columnType: string;
  }
}

function TableCellEx({
  cell,
  setExpandCellOpen,
  expandCellTextRef,
}: {
  cell: Cell<JSONObject, string>;
  setExpandCellOpen: (v: boolean) => void;
  expandCellTextRef: MutableRefObject<string>;
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
          {showExpand && (
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
}

const TableCellExMemo = memo(TableCellEx);

export function DataGrid({ columns, data }: DataGridProps) {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
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
              {row.getVisibleCells().map((cell) => (
                <TableCellExMemo
                  cell={cell}
                  key={cell.id}
                  setExpandCellOpen={setExpandCellOpen}
                  expandCellTextRef={expandCellTextRef}
                />
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </ScrollArea>
  );
}
