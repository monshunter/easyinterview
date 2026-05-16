import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

// D2 visual system: a single global stylesheet imports themes.css and the
// transcribed `ei-global` reset / scrollbar / fadein rules from
// ui-design/src/primitives.jsx.
import "./app/theme/global.css";

import { createAppClient } from "./api/clientFactory";
import { App } from "./app/App";
import { parseInitialRouteHash } from "./app/bootstrapRoute";
import type { LooseRoute } from "./app/normalizeRoute";

declare global {
  interface Window {
    __EASYINTERVIEW_INITIAL_ROUTE__?: LooseRoute;
  }
}

const root = document.getElementById("root");

if (!root) {
  throw new Error("missing #root element for EasyInterview frontend");
}

createRoot(root).render(
  <StrictMode>
    <App
      client={createAppClient()}
      initialRoute={
        window.__EASYINTERVIEW_INITIAL_ROUTE__ ??
        parseInitialRouteHash(window.location.hash)
      }
    />
  </StrictMode>,
);
