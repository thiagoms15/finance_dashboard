import { describe, expect, it } from "vitest";

import { displayName } from "./displayName";

describe("displayName", () => {
  it("uses explicit profile name when available", () => {
    expect(displayName("John Doe", "john.doe@example.com")).toBe("John Doe");
  });

  it("falls back to email local-part in title case", () => {
    expect(displayName("", "thiagoms.15@gmail.com")).toBe("Thiagoms 15");
    expect(displayName(undefined, "ana_maria-silva@example.com")).toBe("Ana Maria Silva");
  });

  it("uses a generic fallback when neither name nor email exists", () => {
    expect(displayName("", "")).toBe("Investor");
    expect(displayName(undefined, undefined)).toBe("Investor");
  });
});
