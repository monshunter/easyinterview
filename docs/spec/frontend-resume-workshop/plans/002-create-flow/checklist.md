# Frontend Resume Workshop Create Flow Checklist

> **版本**: 1.3
> **状态**: completed
> **更新日期**: 2026-07-07

**关联计划**: [plan](./plan.md)

## Phase 1: Create Flow shell

- [x] 1.1 `flow=create` renders `ResumeCreateFlow`.
- [x] 1.2 CreateFlow keeps upload / paste tabs only.
- [x] 1.3 Auth pending action carries route state only and excludes raw resume content.

## Phase 2: Upload / paste registration

- [x] 2.1 Upload path validates file shape, obtains presign data, performs browser PUT, then calls `registerResume`.
- [x] 2.2 Paste path calls `registerResume` with paste payload after non-empty validation.
- [x] 2.3 Side-effect requests include `Idempotency-Key`; polling requests do not.

## Phase 3: Parse polling

- [x] 3.1 `ResumeParseFlow` renders progress state and calls `getResume`.
- [x] 3.2 Ready, failure, timeout and cancel paths are covered by hook/component tests.
- [x] 3.3 Parse stage does not persist raw resume content into URL, pending action, storage or logs.

## Phase 4: Preview confirm

- [x] 4.1 Preview renders parsed structured content.
- [x] 4.2 Confirm uses flat `updateResume` save path through `useResumeSave().overwrite`.
- [x] 4.3 Success navigates back to Resume Workshop list; validation errors stay inline.

## Phase 5: BDD / integration gates

- [x] 5.1 BDD-Gate: E2E.P0.081 create-flow upload/paste happy path is maintained.
- [x] 5.2 BDD-Gate: E2E.P0.082 parse failure / timeout / cancel path is maintained.
- [x] 5.3 BDD-Gate: E2E.P0.083 preview confirm and Home / Workspace CTA handoff are maintained.
- [x] 5.4 `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/create` PASS (10 files, 47 tests; existing React act warning only).
- [x] 5.5 `corepack pnpm --filter @easyinterview/frontend test src/app/screens/resume-workshop/ResumeWorkshopScreen.test.tsx src/app/screens/resume-workshop/fixture-parity.test.ts` PASS (2 files, 16 tests).
