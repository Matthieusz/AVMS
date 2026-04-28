import { z } from "zod";

export const ItemSchema = z.object({
  id: z.number().int().positive(),
  value: z.string(),
  createdAt: z.string().datetime({ offset: true }),
});

export const ListItemsResponseSchema = z.object({
  items: z.array(ItemSchema),
});

export const HealthResponseSchema = z.object({
  status: z.string(),
  message: z.string(),
  timestamp: z.string().datetime({ offset: true }),
  service: z.string(),
});

export type Item = z.infer<typeof ItemSchema>;
export type ListItemsResponse = z.infer<typeof ListItemsResponseSchema>;
export type HealthResponse = z.infer<typeof HealthResponseSchema>;

class ApiError extends Error {
  status: number;
  responseText?: string;

  constructor(message: string, status: number, responseText?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.responseText = responseText;
  }
}

async function apiFetch<T>(path: string, schema: z.ZodType<T>, init?: RequestInit): Promise<T> {
  const response = await fetch(path, init);
  if (!response.ok) {
    const text = await response.text().catch(() => undefined);
    throw new ApiError(`HTTP ${response.status}${text ? `: ${text}` : ""}`, response.status, text);
  }

  const raw = await response.json();
  const parsed = schema.safeParse(raw);
  if (!parsed.success) {
    throw new ApiError(`Invalid response: ${parsed.error.message}`, 0, JSON.stringify(raw));
  }
  return parsed.data;
}

export async function getHealth(): Promise<HealthResponse> {
  return apiFetch("/api/health", HealthResponseSchema);
}

export async function getItems(): Promise<ListItemsResponse> {
  return apiFetch("/api/items", ListItemsResponseSchema);
}

export async function createItem(value: string): Promise<Item> {
  return apiFetch("/api/items", ItemSchema, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ value }),
  });
}

export async function deleteItem(id: number): Promise<void> {
  const response = await fetch(`/api/items/${id}`, { method: "DELETE" });
  if (!response.ok) {
    const text = await response.text().catch(() => undefined);
    throw new ApiError(`HTTP ${response.status}${text ? `: ${text}` : ""}`, response.status, text);
  }
}
