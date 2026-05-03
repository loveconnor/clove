export type ApiErrorPayload = {
  error?: {
    code?: string
    message?: string
  }
}

export class ApiRequestError extends Error {
  status: number
  code?: string
  payload?: ApiErrorPayload

  constructor(status: number, message: string, code?: string, payload?: ApiErrorPayload) {
    super(message)
    this.name = "ApiRequestError"
    this.status = status
    this.code = code
    this.payload = payload
  }
}

export type ApiClientConfig = {
  baseUrl?: string
  fetch?: typeof fetch
  headers?: HeadersInit
  credentials?: RequestCredentials
}

export type ApiClient = {
  baseUrl: string
  fetch: typeof fetch
  headers?: HeadersInit
  credentials?: RequestCredentials
}

export type RequestOptions = {
  method?: "GET" | "POST" | "PUT" | "PATCH" | "DELETE"
  path: string
  body?: unknown
  headers?: HeadersInit
  signal?: AbortSignal
}

const DEFAULT_BASE_URL = "http://localhost:8080"

export function createApiClient(config: ApiClientConfig = {}): ApiClient {
  return {
    baseUrl: config.baseUrl ?? DEFAULT_BASE_URL,
    fetch: config.fetch ?? fetch,
    headers: config.headers,
    credentials: config.credentials ?? "include",
  }
}

export async function request<T>(client: ApiClient, options: RequestOptions): Promise<T> {
  const url = new URL(options.path, client.baseUrl)
  const headers = new Headers(client.headers)

  if (options.headers) {
    const overrides = new Headers(options.headers)
    overrides.forEach((value, key) => {
      headers.set(key, value)
    })
  }

  let body: BodyInit | undefined
  if (options.body !== undefined) {
    headers.set("Content-Type", "application/json")
    body = JSON.stringify(options.body)
  }

  const response = await client.fetch(url.toString(), {
    method: options.method ?? "GET",
    headers,
    body,
    credentials: client.credentials,
    signal: options.signal,
  })

  if (!response.ok) {
    let payload: ApiErrorPayload | undefined
    try {
      payload = (await response.json()) as ApiErrorPayload
    } catch {
      payload = undefined
    }

    throw new ApiRequestError(
      response.status,
      payload?.error?.message ?? "API request failed",
      payload?.error?.code,
      payload
    )
  }

  if (response.status === 204) {
    return undefined as T
  }

  return (await response.json()) as T
}
