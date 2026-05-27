# UAT Resume · Senior Backend Engineer (English)

> Synthetic UAT material. Person and history are fictional; contains no real PII.
> Usage: paste the body below into the resume create flow (ResumeCreateFlow) "paste" entry; or use as a content reference for the Tier-A seeded resume.

---

Name: Lin Hai (fictional)
Target role: Senior Backend Engineer (Go / Distributed Systems)
Location: Shanghai

## Summary

7 years of backend development, 4 of them focused on high-concurrency Go services. Strong in distributed-systems design, async tasks, and event-driven architecture; led several core data pipelines from scratch through reliability governance. Disciplined about idempotency, observability, and data-privacy red lines; drives front-to-back collaboration through OpenAPI contracts.

## Core skills

- Languages & frameworks: Go (expert), SQL, Python (proficient)
- Storage: PostgreSQL (indexing / transaction isolation / query optimization / schema evolution), Redis
- Messaging & async: Kafka, outbox pattern, idempotency keys, retry and compensation, task orchestration
- Cloud-native: Kubernetes, Docker, gRPC, progressive rollout
- Observability: structured logging, Prometheus metrics, OpenTelemetry tracing
- Collaboration: OpenAPI contract design, code review, design review

## Experience

### WeiShu Tech · Senior Backend Engineer (2021.06 — present)

- Led the real-time ingestion service from scratch, supporting 8B events/day with P99 write latency < 120ms.
- Designed reliable event delivery and idempotent consumption on outbox + Kafka, reducing duplicate processing from 0.3% to 0.
- Established SLOs and alerting on critical paths; led 5 incident post-mortems; raised yearly availability from 99.5% to 99.95%.
- Drove cross-team OpenAPI contract standards, cutting front/back integration rework by ~40%.

### QiHang Network · Backend Engineer (2018.07 — 2021.05)

- Owned order and billing microservices; refactored distributed transactions, eliminating oversell and double-charge defects.
- Optimized PostgreSQL slow queries and indexes, dropping core-endpoint P95 latency by 60%.
- Introduced structured logging and metrics, building the first service observability dashboard.

## Projects

### Real-time data replay platform

- Designed a replayable event pipeline supporting time-window replay and compensation for traceable data repair.
- Added backpressure and rate limiting, achieving zero downstream cascading failures at peak load.

## Education

Key University · B.S. in Computer Science (2014 — 2018)
