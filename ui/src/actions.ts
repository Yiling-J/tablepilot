import { fetchEventSource } from "@microsoft/fetch-event-source";
import { JSONObject } from "./json";
import {
    autofillUrl,
    generateUrl,
    modelsUrl,
    rowsUrl,
    tableUrl,
    tablesUrl,
    truncateUrl,
} from "./urls";

export interface TableInfo {
  id: string;
  name: string;
  description: string;
  columns: Array<Column>;
  model: string;
}

export interface GetTablesResponse {
  tables: TableInfo[];
  total: number;
}

export async function getTables(): Promise<GetTablesResponse> {
  const res = await fetch(tablesUrl(), {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    throw new Error("Failed to fetch data");
  }
  return res.json();
}

export async function deleteTable(id: string) {
  const res = await fetch(tableUrl(id), {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    throw new Error("Failed to fetch data");
  }
  return res.status;
}

export type ColumnType = "string" | "number" | "integer" | "boolean" | "array";

export interface Column {
  id: string;
  name: string;
  description: string;
  type: ColumnType;
  fill_mode: string;
}

export async function getRows(id: string): Promise<JSONObject[]> {
  const res = await fetch(rowsUrl(id), {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    throw new Error("Failed to fetch data");
  }
  return res.json().then((v) => v.data);
}

export async function getTable(id: string): Promise<TableInfo> {
  const res = await fetch(tableUrl(id), {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    throw new Error("Failed to fetch data");
  }
  return res.json();
}

interface BaseSource {
  name: string;
  type: string;
}

export interface AiSource extends BaseSource {
  type: "ai";
  prompt: string;
}

export interface ListSource extends BaseSource {
  type: "list";
  options: string[];
}

export interface LinkedSource extends BaseSource {
  type: "linked";
  table: string;
}

export type Source = AiSource | ListSource | LinkedSource;

export interface TableCreateRequest {
  name: string;
  description: string;
  sources: Source[];
  columns: {
    name: string;
    description: string;
    type: string;
    fill_mode: string;
    context_length?: number;
    source?: string;
    random: boolean;
    replacement: boolean;
    repeat: number;
    linked_column: string;
    linked_context_columns: string[];
  }[];
}

export async function createTable(
  request: TableCreateRequest,
): Promise<TableInfo> {
  const res = await fetch(tablesUrl(), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    throw new Error("Failed to fetch data");
  }
  return res.json();
}

export interface ModelList {
  default: string;
  models: string[];
}

export async function getModels(): Promise<ModelList> {
  const res = await fetch(modelsUrl(), {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    throw new Error("Failed to fetch data");
  }
  return res.json();
}

export interface GenerateRequest {
  batch: number;
  count: number;
  temperature: number;
  model: string;
}

export async function generate(
  table: string,
  signal: AbortSignal,
  callback: (data: string) => void,
  { batch, count, temperature, model }: GenerateRequest,
) {
  await fetchEventSource(generateUrl(table), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ batch, count, temperature, model, stream: true }),
    signal: signal,
    onclose() {
      callback("[DONE]");
    },
    onmessage(ev) {
      callback(ev.data);
    },
  });
  callback("[DONE]");
}

export interface AutofillParams {
  columns: string[];
  context_columns: string[];
  offset: number;
}

export interface AutofillRequest {
  genRequest: GenerateRequest;
  autofill: AutofillParams;
}

export async function autofill(
  table: string,
  signal: AbortSignal,
  callback: (data: string) => void,
  { genRequest, autofill }: AutofillRequest,
) {
  await fetchEventSource(autofillUrl(table), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      batch: genRequest.batch,
      count: genRequest.count,
      temperature: genRequest.temperature,
      model: genRequest.model,
      stream: true,
      autofill,
    }),
    signal: signal,
    onclose() {
      callback("[DONE]");
    },
    onmessage(ev) {
      callback(ev.data);
    },
  });
  callback("[DONE]");
}

export async function truncateTable(table: string) {
  const res = await fetch(truncateUrl(table), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: null,
  });
  if (!res.ok) {
    throw new Error("Failed to truncate table");
  }
}

export async function createRows(table: string, rows: JSONObject[]) {
  const res = await fetch(tableUrl(table), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ rows }),
  });
  if (!res.ok) {
    throw new Error("Failed to truncate table");
  }
}
