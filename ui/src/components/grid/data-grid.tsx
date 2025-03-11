import {
    ColumnDef,
    flexRender,
    getCoreRowModel,
    useReactTable,
} from "@tanstack/react-table";

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
import { ScrollArea } from "@radix-ui/react-scroll-area";

interface DataGridProps {
  columns: ColumnDef<JSONObject, string>[];
  data: JSONObject[];
  setHoverCell: (c: string) => void;
}

export function DataGrid({ columns, data, setHoverCell }: DataGridProps) {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <ScrollArea>
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
                <TableCell
                  key={cell.id}
                  className={cn(
                    "border-l border-b last:border-r border-sky-900 max-w-lg max-h-80 whitespace-pre-wrap relative z-0",
                    cell.column.id == "rowIndex" ? "cursor-pointer" : "",
                  )}
                  onMouseEnter={() => {
                    setHoverCell(row.id + "::" + cell.id);
                  }}
                  onMouseLeave={() => setHoverCell("")}
                >
                  {flexRender(cell.column.columnDef.cell, cell.getContext())}
                </TableCell>
              ))}
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </ScrollArea>
  );
}
