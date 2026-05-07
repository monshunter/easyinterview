import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

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
