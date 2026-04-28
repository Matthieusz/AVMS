import { describe, expect, it } from "vite-plus/test";
import { HealthResponseSchema, ItemSchema, ListItemsResponseSchema } from "@/lib/api";

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

describe("ItemSchema", () => {
  it("parses a valid item", () => {
    const data = {
      id: 1,
      value: "hello",
      createdAt: "2026-04-28T12:00:00Z",
    };
    expect(ItemSchema.parse(data)).toEqual(data);
  });

  it("rejects non-positive id", () => {
    expect(() =>
      ItemSchema.parse({ id: 0, value: "x", createdAt: "2026-04-28T12:00:00Z" }),
    ).toThrow();
  });
});

describe("ListItemsResponseSchema", () => {
  it("parses an empty list", () => {
    expect(ListItemsResponseSchema.parse({ items: [] })).toEqual({ items: [] });
  });

  it("parses a populated list", () => {
    const data = {
      items: [
        { id: 1, value: "a", createdAt: "2026-04-28T12:00:00Z" },
        { id: 2, value: "b", createdAt: "2026-04-28T12:01:00Z" },
      ],
    };
    expect(ListItemsResponseSchema.parse(data)).toEqual(data);
  });
});
