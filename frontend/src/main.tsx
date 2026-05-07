import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

// D2 visual system: a single global stylesheet imports themes.css and the
// transcribed `ei-global` reset / scrollbar / fadein rules from
// ui-design/src/primitives.jsx.
import "./app/theme/global.css";

import { EasyInterviewClient } from "./api/generated/client";
import { App } from "./app/App";

const root = document.getElementById("root");

if (!root) {
  throw new Error("missing #root element for EasyInterview frontend");
}

createRoot(root).render(
  <StrictMode>
    <App client={new EasyInterviewClient()} />
  </StrictMode>,
);
