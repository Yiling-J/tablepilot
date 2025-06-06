const address = import.meta.env.VITE_SERVER_ADDRESS;

export function tablesUrl() {
  return `${address}/api/v1/tables`;
}

export function tableUrl(id: string) {
  return `${address}/api/v1/tables/${id}`;
}

export function rowsUrl(id: string) {
  return `${address}/api/v1/tables/${id}/rows`;
}

export function modelsUrl() {
  return `${address}/api/v1/models`;
}

export function generateUrl(id: string) {
  return `${address}/api/v1/generate/tables/${id}`;
}

export function autofillUrl(id: string) {
  return `${address}/api/v1/autofill/tables/${id}`;
}

export function truncateUrl(id: string) {
  return `${address}/api/v1/tables/${id}/truncate`;
}

export function sourcesUrl() {
  return `${address}/api/v1/sources`;
}

export function imageUrl(path: string) {
  return `${address}/api/v1/images/${path}`;
}

export function schemaUrl(id: string) {
  return `${address}/api/v1/tables/${id}/schema`;
}

export function providersUrl() {
  return `${address}/api/v1/providers`;
}

export function providerUrl(id: string) {
  return `${address}/api/v1/providers/${id}`;
}

export function regenerateUrl(id: string) {
  return `${address}/api/v1/regenerate/tables/${id}`;
}

export function importImageUrl() {
  return `${address}/api/v1/image_import/tables`;
}

export function workflowsUrl() {
  return `${address}/api/v1/workflows`;
}

export function getWorkflowUrl(id: string): string {
  return `${address}/api/v1/workflows/${id}`;
}

export function runWorkflowUrl(id: string) {
  return `${address}/api/v1/workflows/${id}/run`;
}

export function datasetsUrl() {
  return `${address}/api/v1/datasets`;
}

export function getDatasetUrl(id: string): string {
  return `${address}/api/v1/datasets/${id}`;
}

export function previewDatasetUrl(id: string) {
  return `${address}/api/v1/datasets/${id}/preview`;
}

export function genDatasetOptionsUrl() {
  return `${address}/api/v1/ai/list_gen`;
}

export function tableDatasetsUrl(id: string) {
  return `${address}/api/v1/tables/${id}/datasets`;
}
