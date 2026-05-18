import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

// D2 visual system: a single global stylesheet imports themes.css and the
// transcribed `ei-global` reset / scrollbar / fadein rules from
// ui-design/src/primitives.jsx.
import "./app/theme/global.css";

import { createAppClient } from "./api/clientFactory";
import { App } from "./app/App";

// Plan 004 §6 Phase 2.1 — initial route resolution lives in routeStore.ts.
// Bootstrap priority:
//   1. window.__EASYINTERVIEW_INITIAL_ROUTE__ (test harness override).
//   2. window.location pathname + search (Browser History canonical URL).
//   3. `#route=...` adapter (preserved for static preview / pixel parity).
//   4. DEFAULT_ROUTE (home).
// `<App />` reads window directly via `useBrowserRoute`; we no longer
// re-derive the initial route here so production parity matches the
// jsdom integration tests.

const root = document.getElementById("root");

if (!root) {
  throw new Error("missing #root element for EasyInterview frontend");
}

createRoot(root).render(
  <StrictMode>
    <App client={createAppClient()} />
  </StrictMode>,
);
