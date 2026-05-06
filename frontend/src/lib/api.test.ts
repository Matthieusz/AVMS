import { describe, expect, it } from "vite-plus/test";
import { HealthResponseSchema, EntrySchema, ListEntriesResponseSchema } from "@/lib/api";

describe("HealthResponseSchema", () => {
  it("parses a valid health response", () => {
    const data = {
      status: "up",
      message: "OK",
      timestamp: "2026-04-28T12:00:00Z",
      service: "avms-api",
    };
    expect(HealthResponseSchema.parse(data)).toEqual(data);
  });

  it("rejects missing fields", () => {
    expect(() => HealthResponseSchema.parse({ status: "up" })).toThrow();
  });
});

describe("EntrySchema", () => {
  it("parses a valid entry", () => {
    const data = {
      id: 1,
      value: "hello",
      createdAt: "2026-04-28T12:00:00Z",
    };
    expect(EntrySchema.parse(data)).toEqual(data);
  });

  it("rejects non-positive id", () => {
    expect(() =>
      EntrySchema.parse({ id: 0, value: "x", createdAt: "2026-04-28T12:00:00Z" }),
    ).toThrow();
  });
});

describe("ListEntriesResponseSchema", () => {
  it("parses an empty list", () => {
    expect(ListEntriesResponseSchema.parse({ entries: [] })).toEqual({ entries: [] });
  });

  it("parses a populated list", () => {
    const data = {
      entries: [
        { id: 1, value: "a", createdAt: "2026-04-28T12:00:00Z" },
        { id: 2, value: "b", createdAt: "2026-04-28T12:01:00Z" },
      ],
    };
    expect(ListEntriesResponseSchema.parse(data)).toEqual(data);
  });
});
