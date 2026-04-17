import { useForm } from "@tanstack/react-form";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";
import { Button } from "./components/ui/button";

interface HealthResponse {
  status: string;
  message: string;
  timestamp: string;
  service: string;
}

interface Item {
  id: number;
  value: string;
  createdAt: string;
}

interface ListItemsResponse {
  items: Item[];
}

async function getHealth(): Promise<HealthResponse> {
  const response = await fetch("/api/health");
  if (!response.ok) {
    throw new Error(`Failed to fetch health: HTTP ${response.status}`);
  }

  return response.json() as Promise<HealthResponse>;
}

async function getItems(): Promise<ListItemsResponse> {
  const response = await fetch("/api/items");
  if (!response.ok) {
    throw new Error(`Failed to fetch items: HTTP ${response.status}`);
  }

  return response.json() as Promise<ListItemsResponse>;
}

async function createItem(value: string): Promise<Item> {
  const response = await fetch("/api/items", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ value }),
  });

  if (!response.ok) {
    throw new Error(`Failed to create item: HTTP ${response.status}`);
  }

  return response.json() as Promise<Item>;
}

async function deleteItem(id: number): Promise<void> {
  const response = await fetch(`/api/items/${id}`, {
    method: "DELETE",
  });

  if (!response.ok) {
    throw new Error(`Failed to delete item: HTTP ${response.status}`);
  }
}

function App() {
  const queryClient = useQueryClient();
  const [deletingItemId, setDeletingItemId] = useState<number | null>(null);

  const healthQuery = useQuery({
    queryKey: ["health"],
    queryFn: getHealth,
    refetchInterval: 30_000,
  });

  const itemsQuery = useQuery({
    queryKey: ["items"],
    queryFn: getItems,
  });

  const createItemMutation = useMutation({
    mutationFn: createItem,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["items"] });
    },
  });

  const deleteItemMutation = useMutation({
    mutationFn: deleteItem,
    onSuccess: () => {
      setDeletingItemId(null);
      void queryClient.invalidateQueries({ queryKey: ["items"] });
    },
    onError: () => {
      setDeletingItemId(null);
    },
  });

  const form = useForm({
    defaultValues: {
      value: "",
    },
    onSubmit: async ({ value }) => {
      await createItemMutation.mutateAsync(value.value.trim());
      form.reset();
    },
  });

  const hasItems = (itemsQuery.data?.items.length ?? 0) > 0;

  return (
    <main className="min-h-screen bg-[radial-gradient(circle_at_top,_hsl(var(--primary)/0.16),_transparent_42%),linear-gradient(180deg,_hsl(var(--background)),_hsl(var(--muted)/0.5))] px-4 py-10 sm:px-6">
      <section className="mx-auto flex w-full max-w-4xl flex-col gap-6">
        <header className="rounded-3xl border border-border/80 bg-card/85 p-6 shadow-sm backdrop-blur sm:p-8">
          <p className="text-xs font-semibold tracking-[0.22em] text-muted-foreground uppercase">
            AVMS Console
          </p>
          <h1 className="mt-2 text-3xl font-semibold tracking-tight text-foreground sm:text-4xl">
            API + SQLite Control Panel
          </h1>
          <p className="mt-3 max-w-2xl text-sm text-muted-foreground sm:text-base">
            Health checks, inserts, and deletes in one Tailwind + shadcn UI.
          </p>
        </header>

        <section className="rounded-3xl border border-border/80 bg-card/90 p-5 shadow-sm sm:p-6">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <h2 className="text-lg font-semibold text-foreground">Service Health</h2>
            {healthQuery.data ? (
              <span
                className={
                  healthQuery.data.status === "up"
                    ? "rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold text-emerald-700"
                    : "rounded-full bg-red-100 px-3 py-1 text-xs font-semibold text-red-700"
                }
              >
                {healthQuery.data.status === "up" ? "Connected" : "Disconnected"}
              </span>
            ) : null}
          </div>

          {healthQuery.isLoading ? (
            <p className="mt-3 text-sm text-muted-foreground">Connecting to API...</p>
          ) : null}

          {healthQuery.isError ? (
            <p className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {healthQuery.error.message}
            </p>
          ) : null}

          {healthQuery.data ? (
            <dl className="mt-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
              <div className="rounded-xl border border-border/80 bg-background/70 p-3">
                <dt className="text-xs font-medium tracking-wide text-muted-foreground uppercase">
                  Status
                </dt>
                <dd className="mt-1 text-sm font-medium text-foreground">
                  {healthQuery.data.status}
                </dd>
              </div>
              <div className="rounded-xl border border-border/80 bg-background/70 p-3">
                <dt className="text-xs font-medium tracking-wide text-muted-foreground uppercase">
                  Service
                </dt>
                <dd className="mt-1 text-sm font-medium text-foreground">
                  {healthQuery.data.service}
                </dd>
              </div>
              <div className="rounded-xl border border-border/80 bg-background/70 p-3">
                <dt className="text-xs font-medium tracking-wide text-muted-foreground uppercase">
                  Message
                </dt>
                <dd className="mt-1 text-sm font-medium text-foreground">
                  {healthQuery.data.message}
                </dd>
              </div>
              <div className="rounded-xl border border-border/80 bg-background/70 p-3">
                <dt className="text-xs font-medium tracking-wide text-muted-foreground uppercase">
                  Timestamp
                </dt>
                <dd className="mt-1 text-sm font-medium text-foreground">
                  {new Date(healthQuery.data.timestamp).toLocaleString()}
                </dd>
              </div>
            </dl>
          ) : null}
        </section>

        <section className="rounded-3xl border border-border/80 bg-card/90 p-5 shadow-sm sm:p-6">
          <h2 className="text-lg font-semibold text-foreground">Data Entries</h2>

          <form
            className="mt-4 grid gap-3 sm:grid-cols-[1fr_auto] sm:items-start"
            onSubmit={(event) => {
              event.preventDefault();
              event.stopPropagation();
              void form.handleSubmit();
            }}
          >
            <form.Field
              name="value"
              validators={{
                onSubmit: ({ value }) => {
                  if (value.trim().length < 2) {
                    return "Please enter at least 2 characters.";
                  }

                  return undefined;
                },
              }}
            >
              {(field) => (
                <div className="flex flex-col gap-2">
                  <label className="text-sm font-medium text-foreground" htmlFor={field.name}>
                    Value
                  </label>
                  <input
                    id={field.name}
                    name={field.name}
                    type="text"
                    placeholder="Enter a value"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(event) => field.handleChange(event.target.value)}
                    disabled={createItemMutation.isPending}
                    className="h-10 rounded-lg border border-input bg-background px-3 text-sm text-foreground placeholder:text-muted-foreground focus-visible:border-ring focus-visible:outline-none focus-visible:ring-3 focus-visible:ring-ring/40"
                  />
                  {field.state.meta.errors[0] ? (
                    <p className="text-sm text-red-700">{field.state.meta.errors[0]}</p>
                  ) : null}
                </div>
              )}
            </form.Field>

            <Button type="submit" className="h-10 sm:mt-7" disabled={createItemMutation.isPending}>
              {createItemMutation.isPending ? "Saving..." : "Add Entry"}
            </Button>
          </form>

          {createItemMutation.isError ? (
            <p className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {createItemMutation.error.message}
            </p>
          ) : null}

          {itemsQuery.isLoading ? (
            <p className="mt-4 text-sm text-muted-foreground">Loading entries...</p>
          ) : null}

          {itemsQuery.isError ? (
            <p className="mt-4 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {itemsQuery.error.message}
            </p>
          ) : null}

          {hasItems ? (
            <ul className="mt-4 space-y-2">
              {itemsQuery.data?.items.map((item) => (
                <li
                  key={item.id}
                  className="flex flex-col gap-2 rounded-xl border border-border/80 bg-background/70 p-3 sm:flex-row sm:items-center sm:justify-between"
                >
                  <div>
                    <p className="text-sm font-medium text-foreground">{item.value}</p>
                    <time className="text-xs text-muted-foreground" dateTime={item.createdAt}>
                      {new Date(item.createdAt).toLocaleString()}
                    </time>
                  </div>

                  <Button
                    type="button"
                    variant="destructive"
                    size="sm"
                    onClick={() => {
                      setDeletingItemId(item.id);
                      void deleteItemMutation.mutateAsync(item.id);
                    }}
                    disabled={deleteItemMutation.isPending && deletingItemId === item.id}
                  >
                    {deleteItemMutation.isPending && deletingItemId === item.id
                      ? "Removing..."
                      : "Remove"}
                  </Button>
                </li>
              ))}
            </ul>
          ) : null}

          {itemsQuery.data && itemsQuery.data.items.length === 0 ? (
            <p className="mt-4 rounded-xl border border-dashed border-border/80 bg-muted/30 px-4 py-6 text-center text-sm text-muted-foreground">
              No entries yet. Add your first one.
            </p>
          ) : null}

          {deleteItemMutation.isError ? (
            <p className="mt-3 rounded-lg border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700">
              {deleteItemMutation.error.message}
            </p>
          ) : null}
        </section>
      </section>
    </main>
  );
}

export default App;
