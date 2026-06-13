import { beforeEach, describe, expect, it, vi } from "vitest";

import { useSessionStore } from "../../features/auth/store";
import { api } from "./client";

const user = {
  id: "user-1",
  name: "Test User",
  email: "user@example.com",
  createdAt: new Date().toISOString(),
  updatedAt: new Date().toISOString()
};

describe("api auth refresh flow", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    window.localStorage.clear();
    useSessionStore.getState().clearAuth();
    useSessionStore.getState().setAuth("expired-token", user, 60);
  });

  it("refreshes the access token once after a 401 and retries the original request", async () => {
    const fetchMock = vi
      .spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(null, { status: 401 }))
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            accessToken: "fresh-token",
            tokenType: "Bearer",
            expiresIn: 900,
            user
          }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" }
          }
        )
      )
      .mockResolvedValueOnce(
        new Response(
          JSON.stringify({
            summary: {
              preferredCurrency: "USD",
              totalInvested: "0",
              currentValue: "0",
              totalProfitLoss: "0",
              dailyGainLoss: "0",
              realizedProfit: "0",
              dividendsReceived: "0"
            },
            positions: []
          }),
          {
            status: 200,
            headers: { "Content-Type": "application/json" }
          }
        )
      );

    const snapshot = await api.portfolio("USD");

    expect(snapshot.positions).toEqual([]);
    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(useSessionStore.getState().token).toBe("fresh-token");
  });

  it("clears the session if refresh also fails", async () => {
    vi.spyOn(globalThis, "fetch")
      .mockResolvedValueOnce(new Response(null, { status: 401 }))
      .mockResolvedValueOnce(new Response(null, { status: 401 }));

    await expect(api.portfolio("USD")).rejects.toThrow("Request failed with 401");
    expect(useSessionStore.getState().token).toBeNull();
  });
});
