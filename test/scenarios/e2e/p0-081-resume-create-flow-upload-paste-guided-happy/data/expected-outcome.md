# P0.081 Expected Outcome

## DOM Anchors

- `resume-create-flow` testid renders for `flow=create` route
- Tab anchors `resume-create-tab-upload/-paste` exist; only the active tab has
  `data-active="true"`; `resume-create-tab-guided` and
  `resume-create-guided-panel` do not exist.
- Upload tab: `resume-create-upload-dropzone`, `resume-create-upload-input`, `resume-create-upload-choose`, `resume-create-upload-selected`
- Paste tab: `resume-create-paste-textarea`, `resume-create-paste-submit`
- Sidebar: `resume-create-sidebar` with `WHAT GETS SAVED` + `WHAT HAPPENS NEXT` text
- Parse stage: `resume-parse-flow` testid, 7 step ticker `resume-parse-step-{0..6}`

## Idempotency / Header

- `createUploadPresign` request carries `Idempotency-Key` in `v1.<unix>.<uuidv7>` format
- `registerResume` request carries `Idempotency-Key` and `Accept-Language` headers
- Idempotency keys are fresh per invocation (no reuse across re-invocations)

## Privacy

- No localStorage / sessionStorage writes during the flow
- No console.log / console.info emitted during the flow
- The signed-URL PUT body is the file blob; the URL/host does not appear in DOM text outside the mock client

## Stage Machine

- After upload PUT + register success, stage transitions to `parsing`
- ParseFlow renders ticker while polling
- On `parseStatus=ready`, stage transitions to `preview`

## Negative Greps

- Source tree under `frontend/src/app/screens/resume-workshop/create/` contains zero retired-module references
- Source tree does not import `ui-design/src/data` or `ui-design/src/screen-resume-workshop`

## Trigger Log Assertions

- `Test Files +\d+ passed` matches in `trigger.log`
- Each of the linked test files is present in the log
