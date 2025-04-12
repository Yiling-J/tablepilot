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

export function imageUrl(tableID: string, path: string) {
  return `${address}/api/v1/tables/${tableID}/images/${path}`;
}
