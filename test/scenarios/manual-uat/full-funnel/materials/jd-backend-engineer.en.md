# UAT JD · Senior Backend Engineer (English)

> Synthetic UAT material. Company and role are fictional; contains no real PII.
> Usage: paste the body below into the Home "paste JD" import entry to trigger parsing.

---

Company: CloudPivot Data
Role: Senior Backend Engineer (Go / Distributed Systems)
Location: Shanghai (Hybrid)
Team: Platform & Infrastructure

## About us

CloudPivot Data builds a real-time data integration and analytics platform for mid-to-large enterprises, processing tens of billions of events per day. Our backend is a fleet of Go microservices on Kubernetes, backed by PostgreSQL, Kafka, and object storage to power high-throughput, observable, replayable data pipelines. We are growing the platform team and looking for a senior engineer who can independently own the design and reliability of core services.

## Responsibilities

- Design, build, and evolve core ingestion and orchestration services, guaranteeing correctness and stability under high concurrency.
- Design idempotent, retryable, observable async tasks and event pipelines (Kafka / outbox / retry and compensation).
- Own capacity, latency, and error budgets on critical paths; establish SLOs and alerting; lead incident triage and post-mortems.
- Align API contracts (OpenAPI) with product, data, and frontend teams to drive stable cross-team interface evolution.
- Enforce data privacy and compliance red lines, ensuring sensitive fields never reach logs or the observability surface.
- Mentor junior and mid-level engineers, participate in code and design reviews, and codify engineering standards.

## Requirements (must-have)

- 5+ years of backend development, 3+ years of production-grade Go services.
- Solid distributed-systems fundamentals: consistency, idempotency, concurrency control, reliable message delivery.
- Strong PostgreSQL: indexing, transaction isolation, query optimization, and schema evolution.
- Familiarity with a message queue (Kafka or NATS) and async task orchestration patterns.
- Hands-on observability practice: structured logging, metrics, distributed tracing.
- Strong API-contract sense; able to collaborate front-to-back via OpenAPI.

## Nice to have

- Experience with data platforms / real-time compute / CDC.
- Familiarity with Kubernetes, cloud-native deployment, and progressive rollout.
- Track record of designing core services from scratch and leading reliability governance.
- Cost/performance trade-off awareness with large-scale efficiency wins.

## Tech stack

Go, PostgreSQL, Kafka, Redis, Kubernetes, gRPC, OpenAPI, OpenTelemetry, S3-compatible object storage.
