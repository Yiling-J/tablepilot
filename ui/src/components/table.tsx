import {
    Column,
    GenerateRequest,
    ModelList,
    TableInfo,
    autofill,
    generate,
    getModels,
    getRows,
    getTable,
    getTableSchema,
    truncateTable,
} from "@/actions";
import { Separator } from "@/components/ui/separator";
import { useCreateTableDialog } from "@/context/create-table";
import { cn } from "@/lib/utils";
import { imageUrl } from "@/urls.tsx";
import { ColumnDef } from "@tanstack/react-table";
import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { AutofillDialog } from "./dialog/autofill-start.tsx";
import { CellTextDialog } from "./dialog/cell-text.tsx";
import { DataGrid } from "./grid/data-grid";
import { Button } from "./ui/button";

import {
    Select,
    SelectContent,
    SelectGroup,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { GearIcon, ReloadIcon } from "@radix-ui/react-icons";
import { download, generateCsv, mkConfig } from "export-to-csv";
import { GridHeader } from "./grid-header.tsx";
import { TablepilotHeader } from "./header.tsx";

import { useTables } from "@/context/tables";
import { JSONObject } from "@/json.ts";
import * as SelectPrimitive from "@radix-ui/react-select";
import { Check } from "lucide-react";

export function TablePage() {
  const { id } = useParams();

  return <Table id={id as string} key={`${id}`} />;
}

const csvConfig = mkConfig({ useKeysAsHeaders: true });

interface TableButton {
  text: string;
  enabled: boolean;
  clickState: string;
  icon: string;
  color: string;
}

interface TableProps {
  id: string;
}

const loading = (
  <div className="w-full h-full flex flex-col pl-0 peer-[[data-state=open]]:lg:pl-[300px] peer-[[data-state=open]]:xl:pl-[300px]">
    <div className="flex flex-row w-full justify-between pt-6 text-xl font-bold">
      <div className="invisible">D</div>
      <div className="flex flex-wrap">
        <Skeleton className="h-6 w-[250px]" />
      </div>
      <div></div>
    </div>

    <div className="my-3 w-full px-4 self-center">
      <Separator />
    </div>

    <div className="pb-3 px-4">
      <Button className="mr-3 bg-green-600 text-white" disabled={true}>
        <span className="cursor-pointer material-symbols-rounded pl-0 pr-2">
          play_circle
        </span>
        Start
      </Button>
    </div>

    <div className="grow overflow-auto">
      <Skeleton className="h-full w-full" />
    </div>

    <div className="pt-3 pb-5 border-t-2 border-t-teal-500 px-4 flex flex-wrap justify-between bg-secondary">
      <div className="flex items-center">
        <p className="align-bottom pr-3 font-semibold	text-slate-500">Rows:</p>
      </div>
      <div>
        <Button className="mr-3" disabled={true}>
          <span className="cursor-pointer material-symbols-rounded pl-0 pr-2">
            download
          </span>
          output.csv
        </Button>
      </div>
    </div>
  </div>
);

function ColumnHeader({
  column,
  onClick,
}: {
  column: Column;
  onClick: () => Promise<void>;
}) {
  const [hoverColumn, setHoverColumn] = useState("");
  return (
    <div
      className="flex content-center text-black dark:text-white text-sm items-center"
      onMouseEnter={() => {
        setHoverColumn(column.id);
      }}
      onMouseLeave={() => setHoverColumn("")}
      onClick={onClick}
    >
      <span className="cursor-pointer material-symbols-rounded pl-2 pr-2 text-base">
        {column.type == "string" && "text_fields"}
        {column.type == "number" && "numbers"}
        {column.type == "integer" && "numbers"}
        {column.type == "array" && "data_array"}
        {column.type == "boolean" && "check"}
        {column.type == "image" && "image"}
      </span>

      <div className="text-base">{column.name}</div>
      {column.id === hoverColumn && <GearIcon className="ml-2" />}
    </div>
  );
}

export function Table({ id }: TableProps) {
  const [rows, setRows] = useState(Array<JSONObject>);
  const [table, setTable] = useState<TableInfo | undefined>(undefined);
  const [models, setModels] = useState<ModelList | undefined>(undefined);
  const [isLoading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [expandCellOpen, setExpandCellOpen] = useState(false);
  const [autofillOpen, setAutofillOpen] = useState(false);
  const [model, setModel] = useState("");
  const [button, setButton] = useState<TableButton>({
    text: "Start",
    enabled: true,
    clickState: "start",
    icon: "play_circle",
    color: "bg-green-600",
  });
  const expandCellTextRef = useRef("");
  const genRef = useRef(false);
  const genRequestRef = useRef({
    batch: 10,
    count: 50,
    temperature: 0.6,
  } as GenerateRequest);
  const abortControllerRef = useRef(new AbortController());
  const columnsRef = useRef([] as ColumnDef<JSONObject, string>[]);
  const modeRef = useRef<"generate" | "autofill">("generate");
  const autofillOffsetRef = useRef(0);
  const { refreshTables } = useTables();
  const { openNewTableDialog, withForm, withTable, withSubmitCallback } =
    useCreateTableDialog();

  const handleEditColumnClick = async () => {
    const schema = await getTableSchema(id);
    withForm(schema);
    withTable(id);
    openNewTableDialog();
    withSubmitCallback(fetchData);
  };

  const fetchData = async () => {
    try {
      const table = await getTable(id);
      // set columns first
      columnsRef.current = [
        {
          accessorKey: "rowIndex",
          header: () => <div></div>,
          meta: { columnType: "integer" },
          size: 10,
          cell: (cell) => {
            return <div className="w-6 h-4">{cell.row.index + 1}</div>;
          },
        },

        ...table.columns.map(
          (e): ColumnDef<JSONObject, string> => ({
            accessorKey: e.id,
            meta: { columnType: e.type.toString() },
            header: () => (
              <ColumnHeader column={e} onClick={handleEditColumnClick} />
            ),
            accessorFn: (row: JSONObject) => {
              const v = row[e.id] as object;
              if (Array.isArray(v)) {
                return v.map((e) => `• ${e}`).join("\n");
              }
              return String(v);
            },

            cell: ({ cell }) => {
              const cellValue = cell.renderValue();
              if (e.type == "image") {
                if (!cell.renderValue()) {
                  return (
                    <div className="w-64 h-64 border flex items-center justify-center rounded">
                      <Skeleton className="w-64 h-64 rounded" />
                    </div>
                  );
                }
                return (
                  <div>
                    <img
                      src={imageUrl(cell.renderValue() as string)}
                      width={256}
                      height={256}
                      className="rounded"
                    />
                  </div>
                );
              }
              return (
                <div>
                  <div className="max-h-80 line-clamp-6">{cellValue}</div>
                </div>
              );
            },
          }),
        ),
      ];
      setTable(table);
      const rows = await getRows(id);
      const vm = [];
      for (const row of rows) {
        vm.push(row);
      }
      await refreshTables();

      const models = await getModels();
      let currentModel = models.default;
      if (table.model) {
        currentModel = table.model;
      }
      setModel(currentModel);
      genRequestRef.current.model = currentModel;
      setModels(models);
      setRows(vm);
      setLoading(false);
    } catch (error) {
      console.error("Error fetching data:", error);
    }
  };

  useEffect(() => {
    fetchData();
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
  }, []);

  const handleGenEvent = (data: string): void => {
    if (data === "[DONE]") {
      setButton({
        text: "Start",
        enabled: true,
        clickState: "start",
        icon: "play_circle",
        color: "bg-green-600",
      });
      setGenerating(false);
      return;
    }
    try {
      const newRows: JSONObject[] = JSON.parse(data).data;
      setRows((old) => {
        return [...old, ...newRows];
      });
    } catch (error) {
      console.error("Error generating data:", error);
    }
  };

  const handleAutofillEvent = (data: string): void => {
    if (data === "[DONE]") {
      setButton({
        text: "Start",
        enabled: true,
        clickState: "start",
        icon: "play_circle",
        color: "bg-green-600",
      });
      setGenerating(false);
      return;
    }
    try {
      const newRows: JSONObject[] = JSON.parse(data).data;
      setRows((old) => {
        return concateRows(old, newRows, autofillOffsetRef.current);
      });
      autofillOffsetRef.current += newRows.length;
    } catch (error) {
      console.error("Error generating data:", error);
    }
  };

  const clickButton = (state: string) => {
    switch (state) {
      case "start": {
        genRequestRef.current.model = model;
        if (modeRef.current === "autofill") {
          autofillOffsetRef.current = 0;
          setAutofillOpen(true);
          return;
        }
        setButton({
          text: "Stop",
          enabled: true,
          clickState: "stop",
          icon: "stop_circle",
          color: "bg-red-600",
        });
        setGenerating(true);
        genRef.current = true;
        abortControllerRef.current = new AbortController();
        generate(
          id,
          abortControllerRef.current.signal,
          handleGenEvent,
          genRequestRef.current,
        );
        break;
      }
      case "stop": {
        genRef.current = false;
        abortControllerRef.current.abort();
        break;
      }
    }
  };

  if (isLoading) {
    return loading;
  }

  if (table === undefined) {
    throw new Error("data undefined");
  }

  const handleExportRows = () => {
    const exported = rows.map((row) => {
      return Object.fromEntries(
        table!.columns.map((header) => {
          const obj = row[header.id]!;
          let cv = "";
          if (Array.isArray(obj)) {
            cv = JSON.stringify(obj);
          } else {
            cv = String(obj);
          }
          return [[header.name], cv];
        }),
      );
    });
    const csv = generateCsv(csvConfig)(exported);
    download(csvConfig)(csv);
  };

  return (
    <div className="grow overflow-hidden h-full flex flex-col pl-0 peer-[[data-state=open]]:lg:pl-[300px] peer-[[data-state=open]]:xl:pl-[300px]">
      <CellTextDialog
        text={expandCellTextRef.current}
        isOpen={expandCellOpen}
        setIsOpen={setExpandCellOpen}
      />
      <AutofillDialog
        isOpen={autofillOpen}
        setIsOpen={setAutofillOpen}
        columns={table.columns}
        onStart={(columns: Column[], contextColumns: Column[]) => {
          setAutofillOpen(false);
          setButton({
            text: "Stop",
            enabled: true,
            clickState: "stop",
            icon: "stop_circle",
            color: "bg-red-600",
          });
          setGenerating(true);
          genRef.current = true;
          abortControllerRef.current = new AbortController();
          autofill(id, abortControllerRef.current.signal, handleAutofillEvent, {
            genRequest: genRequestRef.current,
            autofill: {
              columns: columns.map((c) => c.id),
              context_columns: contextColumns.map((c) => c.id),
              offset: 0,
            },
          });
        }}
      />
      <TablepilotHeader
        title={table.name}
        modeRef={modeRef}
        modeSwitchDisabled={generating}
      />

      <div className="pb-3 px-4 pt-5">
        <div className="flex">
          <Button
            className={cn("mr-3 text-white rounded-sm", button.color)}
            onClick={() => {
              clickButton(button.clickState);
            }}
            disabled={!button.enabled}
          >
            <div className="flex pr-2 justify-center">
              <span className="cursor-pointer material-symbols-rounded">
                {button.icon}
              </span>
            </div>
            {button.text}
          </Button>
          <div className="flex ml-4 border rounded-sm">
            <Select
              value={model}
              disabled={generating}
              onValueChange={async (v) => {
                setModel(v);
              }}
            >
              <SelectTrigger className="w-[180px] ring-0 border-0 focus:ring-offset-0 focus:ring-0 focus:border-0">
                <SelectValue placeholder="Select a model" />
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {models?.models.map((model) => (
                    <SelectPrimitive.Item
                      value={model}
                      key={model}
                      className={cn(
                        "relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none focus:bg-accent focus:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
                        "",
                      )}
                    >
                      <span className="absolute left-2 flex h-3.5 w-3.5 items-center justify-center">
                        <SelectPrimitive.ItemIndicator>
                          <Check className="h-4 w-4" />
                        </SelectPrimitive.ItemIndicator>
                      </span>

                      <div>
                        <SelectPrimitive.ItemText>
                          <p>{model}</p>
                        </SelectPrimitive.ItemText>
                      </div>
                    </SelectPrimitive.Item>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </div>
        </div>

        {generating && (
          <div className="flex pt-2 text-sm text-gray-500 items-center">
            <ReloadIcon className="animate-spin mr-1" />
            <div>Generating...</div>
          </div>
        )}
      </div>
      {!generating && (
        <GridHeader
          genRequestRef={genRequestRef}
          clearData={async () => {
            await truncateTable(id);
            await fetchData();
          }}
        />
      )}
      <div className="scrollbar-thin grow overflow-auto pl-3">
        {table.columns.length > 0 && (
          <DataGrid columns={columnsRef.current} data={rows} />
        )}
      </div>

      <div className="pt-3 pb-5 border-t-2 border-t-teal-500/50 px-4 flex flex-wrap justify-between bg-secondary">
        <div className="flex items-center">
          <p
            className="align-bottom pr-3 font-semibold text-slate-500"
            data-testid="rows-counter"
          >
            Rows: {rows.length}
          </p>
        </div>
        <div>
          <Button className="mr-3" onClick={() => handleExportRows()}>
            <span className="cursor-pointer material-symbols-rounded pl-0 pr-2">
              download
            </span>
            output.csv
          </Button>
        </div>
      </div>
    </div>
  );
}

function concateRows(
  oldRows: JSONObject[],
  newRows: JSONObject[],
  offset: number,
): JSONObject[] {
  if (offset < 0 || offset > oldRows.length) {
    throw new Error("Invalid offset value");
  }

  return [
    ...oldRows.slice(0, offset),
    ...newRows,
    ...oldRows.slice(offset + newRows.length),
  ];
}
