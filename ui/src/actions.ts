import { fetchEventSource } from "@microsoft/fetch-event-source";
import toast from "react-hot-toast";
import { JSONObject } from "./json";
import {
    autofillUrl,
    datasetsUrl,
    generateUrl,
    getWorkflowUrl,
    importImageUrl,
    modelsUrl,
    previewDatasetUrl,
    providerUrl,
    providersUrl,
    regenerateUrl,
    rowsUrl,
    runWorkflowUrl,
    schemaUrl,
    sourcesUrl,
    tableUrl,
    tablesUrl,
    truncateUrl,
    workflowsUrl,
} from "./urls";

export interface TableInfo {
  id: string;
  name: string;
  description: string;
  columns: Array<Column>;
  model: string;
}

// used in workflow only, id will be var path in {{.var}} format
export function tableCreateRequestToTableInfo(
  request: TableCreateRequest,
): TableInfo {
  const convertedColumns: Column[] = request.columns.map((reqCol, _) => {
    return {
      id: reqCol.name,
      name: reqCol.name,
      description: reqCol.description,
      type: reqCol.type as ColumnType,
      fill_mode: reqCol.fill_mode,
    };
  });

  const tableInfo: TableInfo = {
    id: request.name,
    name: request.name,
    description: request.description,
    columns: convertedColumns,
    model: "",
  };

  return tableInfo;
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
    toast.error("Failed to fetch tables");
    throw new Error("Failed to fetch tables");
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
    toast.error("Failed to delete table");
    throw new Error("Failed to delete table");
  }
  return res.status;
}

export type ColumnType =
  | "string"
  | "number"
  | "integer"
  | "boolean"
  | "array"
  | "image";

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
    toast.error("Failed to fetch table rows");
    throw new Error("Failed to fetch table rows");
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
    toast.error("Failed to fetch table details");
    throw new Error("Failed to fetch table details");
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
    toast.error("Failed to create table");
    throw new Error("Failed to create table");
  }
  return res.json();
}

export async function updateTable(
  table: string,
  request: TableCreateRequest,
): Promise<TableInfo> {
  const res = await fetch(tableUrl(table), {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    toast.error("Failed to update table");
    throw new Error("Failed to update table");
  }
  return res.json();
}

export interface ModelListItem {
  name: string;
  image: boolean;
}

export interface ModelList {
  default_model: string;
  default_image_model: string;
  models: ModelListItem[];
}

export async function getModels(): Promise<ModelList> {
  const res = await fetch(modelsUrl(), {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    toast.error("Failed to fetch models");
    throw new Error("Failed to fetch models");
  }
  return res.json();
}

export interface GenerateRequest {
  batch: number;
  count: number;
  temperature: number;
  model: string;
  image_model: string;
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
    openWhenHidden: true,
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
  prompt: string;
  rows: string[];
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
    openWhenHidden: true,
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

export async function regenerate(
  table: string,
  signal: AbortSignal,
  callback: (data: string) => void,
  { genRequest, autofill }: AutofillRequest,
) {
  await fetchEventSource(regenerateUrl(table), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    openWhenHidden: true,
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
    toast.error("Failed to truncate table");
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
    toast.error("Failed to create rows");
    throw new Error("Failed to create rows");
  }
}

export interface SourceData {
  name: string;
  data: JSONObject;
  columns: string[];
}

export async function getSources(): Promise<SourceData[]> {
  const res = await fetch(sourcesUrl(), {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    toast.error("Failed to fetch sources");
    throw new Error("Failed to fetch sources");
  }
  return res.json().then((v) => v.sources);
}

export async function getTableSchema(
  table: string,
): Promise<TableCreateRequest> {
  const res = await fetch(schemaUrl(table), {
    method: "GET",
  });
  if (!res.ok) {
    toast.error("Failed to fetch table schema");
    throw new Error("Failed to fetch table schema");
  }
  return res.json();
}

export interface Model {
  model: string;
  alias: string;
  max_tokens: number;
  rpm: number;
  image: boolean;
}

export type ProviderType =
  | "OpenAI"
  | "Gemini"
  | "Anthropic"
  | "OpenRouter"
  | "OpenAI-Compatible"
  | string;

export const ProviderTypeOptions = [
  "OpenAI",
  "Gemini",
  "Anthropic",
  "OpenRouter",
  "OpenAI-Compatible",
];

export interface Provider {
  id: number;
  name: string;
  type: string;
  key: ProviderType;
  base_url: string;
  models: Model[];
  editable: boolean;
  enabled: boolean;
}

export async function getProviders(): Promise<Provider[]> {
  const res = await fetch(providersUrl(), {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!res.ok) {
    toast.error("Failed to fetch providers");
    throw new Error("Failed to fetch providers");
  }

  return res.json();
}

export async function deleteProvider(id: string) {
  const res = await fetch(providerUrl(id), {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    toast.error("Failed to delete provider");
    throw new Error("Failed to delete provider");
  }
  return res.status;
}

export async function createProvider(provider: Provider): Promise<TableInfo> {
  const res = await fetch(providersUrl(), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(provider),
  });
  if (!res.ok) {
    toast.error("Failed to create provider");
    throw new Error("Failed to create provider");
  }
  return res.json();
}

export async function updateProvider(
  id: string,
  provider: Provider,
): Promise<TableInfo> {
  const res = await fetch(providerUrl(id), {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(provider),
  });
  if (!res.ok) {
    toast.error("Failed to update provider");
    throw new Error("Failed to update provider");
  }
  return res.json();
}

export interface ImportRequest {
  data: string;
  prompt: string;
  model: string;
  table: string;
  name: string;
  truncate: boolean;
}

export async function importImage(req: ImportRequest): Promise<string> {
  const res = await fetch(importImageUrl(), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(req),
  });
  if (!res.ok) {
    toast.error("Failed to import image");
    throw new Error("Failed to import image");
  }
  return res.json().then((v) => v.id);
}

export interface WorkflowInfo {
  id: string;
  name: string;
  description: string;
}

export interface GetWorkflowsResponse {
  workflows: WorkflowInfo[];
  total: number;
}

export async function getWorkflows(): Promise<GetWorkflowsResponse> {
  const res = await fetch(workflowsUrl(), {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    toast.error("Failed to fetch workflows");
    throw new Error("Failed to fetch workflows");
  }
  return res.json();
}

export type WorkflowVariableType = "string" | "number" | "integer" | "file";

export type WorkflowStepType =
  | "UserInput"
  | "CreateTable"
  | "Import"
  | "CreateColumn"
  | "DeleteColumn"
  | "Generate"
  | "Autofill"
  | "ExportTable"
  | "DeleteTable";

export interface WorkflowVariable {
  name: string;
  type: WorkflowVariableType;
  default_value: string | number;
  options: (string | number)[];
}

export interface UserInputStepPayload {
  variables: WorkflowVariable[];
}

export interface CreateTableStepPayload {
  request: TableCreateRequest;
  on_exists: string;
}

export interface DeleteTableStepPayload {
  table: string;
}

export interface CreateColumnStepPayload {
  table: string;
  name: string;
  description: string;
  type: string;
}

export interface DeleteColumnStepPayload {
  table: string;
  column: string;
}

export interface ImportDataStepPayload {
  table: string;
  truncate: boolean;
  name: string;
  file: string;
  prompt: string;
}

export interface GenerateStepPayload {
  table: string;
  batch: number;
  count: number;
}

export interface AutofillStepPayload {
  table: string;
  batch: number;
  count: number;
  columns: string[];
  context_columns: string[];
  prompt: string;
}

export interface ExportStepPayload {
  table: string;
}

interface WorkflowStepPayloadMap {
  UserInput: UserInputStepPayload;
  CreateTable: CreateTableStepPayload;
  Import: ImportDataStepPayload;
  CreateColumn: CreateColumnStepPayload;
  DeleteColumn: DeleteColumnStepPayload;
  Generate: GenerateStepPayload;
  Autofill: AutofillStepPayload;
  DeleteTable: DeleteTableStepPayload;
  ExportTable: ExportStepPayload;
}

export type TypedWorkflowStep = {
  [K in WorkflowStepType]: {
    type: K;
    payload: K extends keyof WorkflowStepPayloadMap
      ? WorkflowStepPayloadMap[K]
      : never; // `never` for unmapped means they shouldn't exist or are an error
    status?: string;
  };
}[WorkflowStepType]; // This gets a union of all possible objects.

export interface Workflow {
  id: string;
  name: string;
  description: string;
  variables: WorkflowVariable[];
  steps: TypedWorkflowStep[];
}

export async function getWorkflow(id: string): Promise<Workflow> {
  const res = await fetch(getWorkflowUrl(id), {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!res.ok) {
    toast.error("Failed to fetch workflow");
    throw new Error("Failed to fetch workflow");
  }

  return res.json();
}

export async function runWorkflow(
  workflow: string,
  signal: AbortSignal,
  callback: (data: string) => void,
  temperature: number,
  model: string,
  imageModel: string,
  variables: JSONObject,
) {
  await fetchEventSource(runWorkflowUrl(workflow), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      temperature,
      model,
      image_model: imageModel,
      variables,
    }),
    signal: signal,
    openWhenHidden: true,
    onclose() {
      callback("[DONE]");
    },
    onmessage(ev) {
      callback(ev.data);
    },
  });
  callback("[DONE]");
}

export async function createWorkflow(request: Workflow): Promise<string> {
  const res = await fetch(workflowsUrl(), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    toast.error("Failed to create workflow");
    throw new Error("Failed to create workflow");
  }
  return res.json().then((v) => v.id);
}

export async function updateWorkflow(
  id: string,
  request: Workflow,
): Promise<string> {
  const res = await fetch(getWorkflowUrl(id), {
    method: "PATCH",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(request),
  });
  if (!res.ok) {
    toast.error("Failed to update workflow");
    throw new Error("Failed to update workflow");
  }
  return res.json().then((v) => v.id);
}

export async function deleteWorkflow(id: string) {
  const res = await fetch(getWorkflowUrl(id), {
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
    },
  });

  if (!res.ok) {
    toast.error("Failed to delete workflow");
    throw new Error("Failed to delete workflow");
  }
}

export type DatasetType = "list" | "csv";

export interface DatasetInfo {
  id: string;
  name: string;
  type: DatasetType;
  description: string;
}

export interface GetDatasetsResponse {
  datasets: DatasetInfo[];
  total: number;
}

export async function getDatasets(): Promise<GetDatasetsResponse> {
  const res = await fetch(datasetsUrl(), {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    toast.error("Failed to fetch workflows");
    throw new Error("Failed to fetch workflows");
  }
  return res.json();
}

export interface CreateDatasetRequest {
  name: string;
  description: string;
  type: DatasetType;
  data: string[];
  files: File[];
}

export async function createDataset(
  req: CreateDatasetRequest,
): Promise<string> {
  const formData = new FormData();
  formData.append("name", req.name);
  formData.append("description", req.description);
  formData.append("type", req.type);

  for (let i = 0; i < req.files.length; i++) {
    formData.append("files", req.files[i]);
  }

  const res = await fetch(datasetsUrl(), {
    method: "POST",
    body: formData,
  });
  if (!res.ok) {
    toast.error("Failed to create dataset");
    throw new Error("Failed to create dataset");
  }
  return res.json();
}

export interface DatasetPreviewResponse {
  type: DatasetType;
  rows: JSONObject[];
  data: string[];
}

export async function previewDataset(
  id: string,
): Promise<DatasetPreviewResponse> {
  const res = await fetch(previewDatasetUrl(id), {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!res.ok) {
    toast.error("Failed to fetch workflows");
    throw new Error("Failed to fetch workflows");
  }
  return res.json();
}
