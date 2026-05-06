import { z } from "zod";

export const EntrySchema = z.object({
  id: z.number().int().positive(),
  value: z.string(),
  createdAt: z.string().datetime({ offset: true }),
});

export const ListEntriesResponseSchema = z.object({
  entries: z.array(EntrySchema),
});

export const HealthResponseSchema = z.object({
  status: z.string(),
  message: z.string(),
  timestamp: z.string().datetime({ offset: true }),
  service: z.string(),
});

export type Entry = z.infer<typeof EntrySchema>;
export type ListEntriesResponse = z.infer<typeof ListEntriesResponseSchema>;
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

export async function getEntries(): Promise<ListEntriesResponse> {
  return apiFetch("/api/entries", ListEntriesResponseSchema);
}

export async function createEntry(value: string): Promise<Entry> {
  return apiFetch("/api/entries", EntrySchema, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ value }),
  });
}

export async function deleteEntry(id: number): Promise<void> {
  const response = await fetch(`/api/entries/${id}`, { method: "DELETE" });
  if (!response.ok) {
    const text = await response.text().catch(() => undefined);
    throw new ApiError(`HTTP ${response.status}${text ? `: ${text}` : ""}`, response.status, text);
  }
}
