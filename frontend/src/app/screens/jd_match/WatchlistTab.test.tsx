// @vitest-environment jsdom
import { describe, expect, it, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import type { ReactNode } from "react";

import { DisplayPreferencesProvider } from "../../display/DisplayPreferencesProvider";
import type {
  MarketSignal,
  WatchlistItem,
} from "../../../api/generated/types";

import { WatchlistTab } from "./WatchlistTab";

function makeItem(overrides: Partial<WatchlistItem> = {}): WatchlistItem {
  return {
    id: "wl-1",
    linkedJobMatchId: "jm-1",
    label: null,
    title: "Senior FE",
    company: "Acme",
    tone: "ok",
    addedAt: "2026-05-08T10:00:00Z",
    change: null,
    ...overrides,
  };
}

function makeSignal(overrides: Partial<MarketSignal> = {}): MarketSignal {
  return {
    k: "Postings · 7d",
    v: "182",
    d: "+12% vs prior week",
    tone: "ok",
    ...overrides,
  };
}

function wrap(node: ReactNode, lang: "zh" | "en" = "en") {
  return render(
    <DisplayPreferencesProvider initial={{ lang }}>
      {node}
    </DisplayPreferencesProvider>,
  );
}

describe("WatchlistTab — view-model parity (item 5.1)", () => {
  it("renders root testid + heading", () => {
    wrap(
      <WatchlistTab
        items={[]}
        loading={false}
        error={null}
        signals={[]}
        signalsLoading={false}
        signalsError={null}
        onChevron={() => undefined}
      />,
    );
    expect(screen.getByTestId("jdmatch-watchlist-tab")).toBeInTheDocument();
    expect(
      screen.getByTestId("jdmatch-watchlist-refresh-footer"),
    ).toBeInTheDocument();
  });

  it("renders empty state when items is [] and not loading", () => {
    wrap(
      <WatchlistTab
        items={[]}
        loading={false}
        error={null}
        signals={[]}
        signalsLoading={false}
        signalsError={null}
        onChevron={() => undefined}
      />,
    );
    expect(screen.getByTestId("jdmatch-watchlist-empty")).toBeInTheDocument();
  });

  it("renders error state when error is set", () => {
    wrap(
      <WatchlistTab
        items={[]}
        loading={false}
        error={new Error("boom")}
        signals={[]}
        signalsLoading={false}
        signalsError={null}
        onChevron={() => undefined}
      />,
    );
    expect(screen.getByTestId("jdmatch-watchlist-error")).toBeInTheDocument();
  });

  it("renders one row per item with stable testid and tone data-attr", () => {
    const items = [
      makeItem({ id: "wl-a", tone: "ok" }),
      makeItem({ id: "wl-b", tone: "warn" }),
      makeItem({ id: "wl-c", tone: "muted" }),
    ];
    wrap(
      <WatchlistTab
        items={items}
        loading={false}
        error={null}
        signals={[]}
        signalsLoading={false}
        signalsError={null}
        onChevron={() => undefined}
      />,
    );
    expect(screen.getByTestId("jdmatch-watchlist-item-wl-a")).toBeInTheDocument();
    expect(
      screen.getByTestId("jdmatch-watchlist-item-wl-b").getAttribute("data-tone"),
    ).toBe("warn");
    expect(
      screen.getByTestId("jdmatch-watchlist-item-wl-c").getAttribute("data-tone"),
    ).toBe("muted");
  });

  it("renders change span only when change is truthy", () => {
    const items = [
      makeItem({ id: "wl-a", change: null }),
      makeItem({ id: "wl-b", change: "JD updated" }),
    ];
    wrap(
      <WatchlistTab
        items={items}
        loading={false}
        error={null}
        signals={[]}
        signalsLoading={false}
        signalsError={null}
        onChevron={() => undefined}
      />,
    );
    expect(
      screen.queryByTestId("jdmatch-watchlist-item-wl-a-change"),
    ).toBeNull();
    expect(
      screen.getByTestId("jdmatch-watchlist-item-wl-b-change"),
    ).toBeInTheDocument();
  });

  it("chevron click invokes onChevron with the clicked WatchlistItem", () => {
    const items = [makeItem({ id: "wl-a", linkedJobMatchId: "jm-target" })];
    const onChevron = vi.fn();
    wrap(
      <WatchlistTab
        items={items}
        loading={false}
        error={null}
        signals={[]}
        signalsLoading={false}
        signalsError={null}
        onChevron={onChevron}
      />,
    );
    fireEvent.click(
      screen.getByTestId("jdmatch-watchlist-item-wl-a-chevron"),
    );
    expect(onChevron).toHaveBeenCalledTimes(1);
    expect(onChevron.mock.calls[0]![0]).toEqual(
      expect.objectContaining({ id: "wl-a", linkedJobMatchId: "jm-target" }),
    );
  });

  it("renders 4 market signals with positional testids and signal-key data-attr", () => {
    const signals: MarketSignal[] = [
      makeSignal({ k: "Postings · 7d", v: "182", tone: "ok" }),
      makeSignal({ k: "Median comp · senior", v: "82 LPA", tone: "warn" }),
      makeSignal({ k: "Remote share", v: "41%", tone: "ok" }),
      makeSignal({ k: "Avg time-to-hire", v: "32 days", tone: "muted", d: null }),
    ];
    wrap(
      <WatchlistTab
        items={[]}
        loading={false}
        error={null}
        signals={signals}
        signalsLoading={false}
        signalsError={null}
        onChevron={() => undefined}
      />,
    );
    for (let i = 0; i < 4; i++) {
      expect(
        screen.getByTestId(`jdmatch-market-signal-${i}`),
      ).toBeInTheDocument();
    }
    expect(
      screen.getByTestId("jdmatch-market-signal-3-fallback"),
    ).toBeInTheDocument();
    expect(
      screen
        .getByTestId("jdmatch-market-signal-1")
        .getAttribute("data-signal-tone"),
    ).toBe("warn");
  });

  it("market signals error surface renders when signalsError is set", () => {
    wrap(
      <WatchlistTab
        items={[]}
        loading={false}
        error={null}
        signals={[]}
        signalsLoading={false}
        signalsError={new Error("boom")}
        onChevron={() => undefined}
      />,
    );
    expect(
      screen.getByTestId("jdmatch-market-signals-error"),
    ).toBeInTheDocument();
  });

  it("partial-data: renders the signals it has and shows a fallback dash for missing d", () => {
    const signals: MarketSignal[] = [
      makeSignal({ k: "Postings · 7d", v: "182", d: null, tone: "ok" }),
      makeSignal({ k: "Median comp · senior", v: "—", d: null, tone: "muted" }),
    ];
    wrap(
      <WatchlistTab
        items={[]}
        loading={false}
        error={null}
        signals={signals}
        signalsLoading={false}
        signalsError={null}
        onChevron={() => undefined}
      />,
    );
    expect(screen.getByTestId("jdmatch-market-signal-0")).toBeInTheDocument();
    expect(screen.getByTestId("jdmatch-market-signal-1")).toBeInTheDocument();
    // both lack d, so both fallbacks render
    expect(
      screen.getByTestId("jdmatch-market-signal-0-fallback"),
    ).toBeInTheDocument();
    expect(
      screen.getByTestId("jdmatch-market-signal-1-fallback"),
    ).toBeInTheDocument();
  });
});
