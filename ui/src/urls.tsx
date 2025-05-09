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
