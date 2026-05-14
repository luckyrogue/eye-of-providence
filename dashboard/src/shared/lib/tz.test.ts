import { describe, expect, it } from "vitest";
import { UNIQUE_TIMEZONES, formatDate, formatTime } from "./tz";
describe("tz utilities", () => {
  it("UNIQUE_TIMEZONES is deduplicated by value", () => {
    const values = UNIQUE_TIMEZONES.map((t) => t.value);
    expect(new Set(values).size).toBe(values.length);
  });
  it("formatDate renders an ISO timestamp in the requested zone", () => {
    const out = formatDate("2025-01-02T10:30:00Z", "Europe/Moscow");
    expect(out).toMatch(/2025/);
    expect(out).toMatch(/13:30/);
  });
  it("formatTime renders only HH:MM:SS in the requested zone", () => {
    const out = formatTime("2025-01-02T10:30:00Z", "Europe/Moscow");
    expect(out).toMatch(/13:30:00/);
  });
});
