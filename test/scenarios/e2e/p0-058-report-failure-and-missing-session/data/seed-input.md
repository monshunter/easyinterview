# Seed input

- Route variants:
  - `report?reportStatus=failed&errorCode=AI_PROVIDER_TIMEOUT&sessionId=…&reportId=…`
  - `report?reportId=…` (missing sessionId)
  - `report?sessionId=…&reportId=…` with backend 404 for the cross-user case
  - `generating?reportId=…&sessionId=…` with persistent `report-generating` fixture for timeout
