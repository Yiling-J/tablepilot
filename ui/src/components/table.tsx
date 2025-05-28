import {
    Column,
    GenerateRequest,
    TableInfo,
    autofill,
    generate,
    getRows,
    getTable,
    getTableSchema,
    regenerate,
    truncateTable,
} from "@/actions";
import { Separator } from "@/components/ui/separator";
import { useCreateTableDialog } from "@/context/create-table";
import { cn } from "@/lib/utils";
import { imageUrl } from "@/urls.tsx";
import { ColumnDef } from "@tanstack/react-table";
import { save } from "@tauri-apps/plugin-dialog";
import { writeTextFile } from "@tauri-apps/plugin-fs";
import { useEffect, useRef, useState } from "react";
import { useParams } from "react-router-dom";
import { AutofillDialog } from "./dialog/autofill-start.tsx";
import { CellTextDialog } from "./dialog/cell-text.tsx";
import { DataGrid } from "./grid/data-grid";
import { Button } from "./ui/button";

import { Skeleton } from "@/components/ui/skeleton";
import { GearIcon, ReloadIcon } from "@radix-ui/react-icons";
import { asString, download, generateCsv, mkConfig } from "export-to-csv";
import { GridHeader } from "./grid-header.tsx";
import { TablepilotHeader } from "./header.tsx";

import { useTables } from "@/context/tables";
import { JSONObject } from "@/json.ts";
import { info } from "@tauri-apps/plugin-log";
import { RegenerateDialog } from "./dialog/regenerate.tsx";
import { ModelSelector } from "./model-selector.tsx";

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
  const [isLoading, setLoading] = useState(true);
  const [generating, setGenerating] = useState(false);
  const [expandCellOpen, setExpandCellOpen] = useState(false);
  const [autofillOpen, setAutofillOpen] = useState(false);
  const [model, setModel] = useState("");
  const [imageModel, setImageModel] = useState("");
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
  const [regenerateRows, setRegenerateRows] = useState([] as string[]);
  const [loadingRows, setLoadingRows] = useState([] as string[]);
  const [regenerateDialogOpen, setRegenerateDialogOpen] = useState(false);
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

  const imageColumnExists =
    table?.columns.find((c) => c.type === "image") !== undefined;

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

  function updateRows(
    oldRows: JSONObject[],
    newRows: JSONObject[],
  ): JSONObject[] {
    const newRowsMap = new Map(
      newRows.map((row) => [row.__id__ as string, row]),
    );
    setLoadingRows((old) => old.filter((row) => !newRowsMap.has(row)));

    return oldRows.map((row) => {
      const id = row.__id__ as string;
      if (newRowsMap.has(id)) {
        return newRowsMap.get(id)!;
      }
      return row;
    });
  }

  const handleRegenerateEvent = (data: string): void => {
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
        return updateRows(old, newRows);
      });
    } catch (error) {
      console.error("Error generating data:", error);
    }
  };

  const clickButton = (state: string) => {
    switch (state) {
      case "start": {
        genRequestRef.current.model = model;
        genRequestRef.current.image_model = imageModel;
        if (regenerateRows.length > 0) {
          setRegenerateDialogOpen(true);
          return;
        }
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

  const handleExportRows = async () => {
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

    if ("__TAURI_INTERNALS__" in window) {
      const path = await save({
        filters: [
          {
            name: "output.csv",
            extensions: ["csv"],
          },
        ],
      });
      if (path) {
        try {
          await writeTextFile(path, asString(csv));
        } catch (e) {
          info(e as string);
        }
        info("CSV file saved successfully to:" + path);
      } else {
        info("Save operation cancelled by user.");
      }
    } else {
      download(csvConfig)(csv);
    }
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
        onStart={(
          columns: string[],
          contextColumns: string[],
          prompt: string,
        ) => {
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
              columns: columns,
              context_columns: contextColumns,
              offset: 0,
              rows: [],
              prompt: prompt,
            },
          });
        }}
      />
      <RegenerateDialog
        open={regenerateDialogOpen}
        onOpenChange={setRegenerateDialogOpen}
        tableInfo={table}
        onRegenerate={(config) => {
          const rows = [...regenerateRows];
          setLoadingRows([...regenerateRows]);
          setRegenerateRows([]);
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
          regenerate(
            id,
            abortControllerRef.current.signal,
            handleRegenerateEvent,
            {
              genRequest: genRequestRef.current,
              autofill: {
                columns: config.columnsToRegenerate,
                context_columns: [],
                offset: 0,
                rows: rows,
                prompt: config.prompt,
              },
            },
          );
        }}
      />
      <TablepilotHeader
        title={table.name}
        modeRef={modeRef}
        modeSwitchDisabled={generating}
      />

      <div className="pb-3 px-4 pt-5">
        <div className="flex items-center">
          <Button
            className={cn("mr-3 text-white rounded-sm", button.color)}
            onClick={() => {
              clickButton(button.clickState);
            }}
            disabled={!button.enabled || (model === "" && imageModel === "")}
          >
            <div className="flex pr-2 justify-center">
              <span className="cursor-pointer material-symbols-rounded">
                {button.icon}
              </span>
            </div>
            {button.text}
          </Button>

          <div>
            <ModelSelector
              hasImageColumn={imageColumnExists}
              generating={generating}
              selectModel={(v) => {
                setModel(v);
                genRequestRef.current.model = v;
              }}
              selectImageModel={(v) => {
                setImageModel(v);
                genRequestRef.current.image_model = v;
              }}
            />
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
          <DataGrid
            columns={columnsRef.current}
            data={rows}
            selectedRows={regenerateRows}
            loading={loadingRows}
            onRowSelectChange={(row, v) => {
              const current = regenerateRows.length;
              if (v) {
                setRegenerateRows([...regenerateRows, row]);
              } else {
                setRegenerateRows(regenerateRows.filter((v) => v !== row));
              }
              if (current === 0 && v) {
                setButton({
                  text: "Regenerate",
                  enabled: true,
                  clickState: "start",
                  icon: "play_circle",
                  color: "bg-green-600",
                });
              }
              if (current === 1 && !v) {
                setButton({
                  text: "Start",
                  enabled: true,
                  clickState: "start",
                  icon: "play_circle",
                  color: "bg-green-600",
                });
              }
            }}
          />
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
