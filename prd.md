# Product Requirements Document — **“Un-tie.me code”**

---

## 1  Overview  
Un-tie.me code is a web application that enables a single software architect to specify, generate, and iteratively evolve a complete cloud-native product with the assistance of hierarchical AI agents. The workbench combines an opinionated, XP-inspired planning flow with automated code generation, CI/CD, and runtime deployment on Google Cloud (Cloud Run or GKE). All build artefacts are containerised; the dev→prod path is guarded by managed pipelines and human-in-the-loop reviews.

Target user Experienced individual contributor who owns product definition, architecture, and delivery; they need deterministic traceability from requirement → architecture → code → deployment without managing multiple unintegrated tools.

Value proposition Reduces cognitive load, enforces modular design discipline, and supplies agentic automation that scales the architect’s output while preserving fine-grained control.

---

## 2  Core Features  
| # | Name | Purpose | Key Elements |
|---|------|---------|--------------|
|0|Authentication|Secure access & device trust|SSO/MFA, Google Identity |
|1|Project Definition Workspace|Capture “source of truth”|Narrative overview, feature-iteration matrix, stack spec, auto-generated **`prd.md`** |
|2|Architectural Canvas|Visual, C4-style architecture|Zoomable C4 diagram; component side-panels with metrics & agent hooks |
|3|Story-Flow Board|XP-flavoured Kanban|Epics → Stories → Features; WIP limits; cycle-time analytics |
|4|Task Monitoring Hub|Observe agent tasks|DAG view of AI and CI tasks with status & critical path |
|5|Human Review Queue|Gatekeeper for quality|Diff viewer; approve/redo controls |
|6|Conversational Design Assistant|Natural-language control surface|Persistent chat summarised into structured artefacts |

---

## 3  User Experience  
1. **Login** → dashboard of projects.  
2. **Create Project** → fill narrative; enumerate UX features row-wise and MVP→v_n columns; add tech-stack snippets; click “Generate PRD” to preview and commit.  
3. **Architectural Canvas** auto-renders; architect clicks a component, edits guidance Markdown, or spawns a scoped code-mod task.  
4. **Story-Flow Board** shows work items; drag to “Ready-AI” to queue automated implementation.  
5. **Task Hub** depicts running agents and CI jobs; architect inspects artifacts or aborts cascades.  
6. **Review Queue** surfaces items needing human judgement; actions feed back to Kanban status.  
7. **Deployment** button triggers Cloud Build → Artifact Registry → Cloud Run/GKE rollout; status visible inline.

Front-end stack: **HTMX** for hypermedia interactions, **Tailwind CSS** for utility-first styling, minimal custom JS.

---

## 4  Technical Architecture  

### 4.1 High-Level Components  
* **Browser client** (HTMX + Tailwind)  
* **API-Gateway & App-Server** – Go + **Gin** REST endpoints  
* **Agent Orchestration Layer** – Genkit Go framework instantiates LLM-powered agents; message-oriented via the **Agent-to-Agent (A2A) protocol** :contentReference[oaicite:0]{index=0}  
* **Context Management Service** – supplies agents with least-necessary project slices; implemented with Redis plus doc-store fallback  
* **Orchestration Engine** – Temporal Cloud for fault-tolerant, long-running orchestration :contentReference[oaicite:1]{index=1}  
* **CI/CD** – Cloud Build triggers on Git tags; Docker images pushed to Artifact Registry then deployed via Cloud Deploy to Cloud Run or GKE  
* **Data Persistence** – Cloud SQL (metadata), Cloud Storage (artefacts), Firestore (chat transcripts)  
* **Observability** – Cloud Logging + Cloud Trace; Temporal Web UI (post-migration)

### 4.2 Container & Pipeline  
* **Dockerfile.dev** – hot-reload Gin + static HTMX assets  
* **Dockerfile.prod** – multi-stage build → minimal distroless image  
* **cloudbuild.yaml** – lint → test → image build → helm-chart render → gated deploy  
* **Helm chart** (GKE) or **Cloud Run service-yaml** (Cloud Run) parameterised by env.

### 4.3 Security  
* IAM-scoped service accounts for build & runtime  
* Secrets in Secret Manager; injected at deploy  
* Temporal mTLS when adopted.

---

## 5  Development Roadmap  

| Phase | Milestone | Scope |
|-------|-----------|-------|
|0|Spike|Single-project CRUD, manual Docker build, stub chat |
|1|MVP|Auth, Project Workspace, CI/CD to Cloud Run |
|2|Architecture Canvas & Genkit agents|Initial DAG executor; A2A schema v0 |
|3|Task Hub & Review Queue|Agent-generated tasks; human-gate wiring |
|4|Temporal Migration|Replace DAG executor; add retries, schedules |
|5|GKE Option & Autoscaling|Helm-based deploy; HPA tuned |
|6|Marketplace & Templates|Shareable PRD/arch blueprints |

---

## 6  Logical Dependency Chain  

1. **Docker** images →  
2. **Cloud Build** pipeline →  
3. **Artifact Registry** artefacts →  
4. **Cloud Run/GKE** runtime →  
5. **Gin API** serves HTMX →  
6. **Genkit agents** orchestrated via **A2A**; rely on Context Svc →  
7. **Workflow engine** (DAG → Temporal) schedules agent & CI jobs; exposes status to UI.

---

## 7  Risks & Mitigations  

| Risk | Impact | Mitigation |
|------|--------|-----------|
|Agent hallucination ⇒ wrong code|Deployment breakage|Strict component-scoping, unit-test auto-generation, mandatory human reviews|
|Context explosion ⇒ latency|Agent cost & lag|Sliding-window context index; vector-store relevance filter|
|Early DAG runner lacks durability|Orchestration failure|Rapid migration plan to Temporal; idempotent task design|
|GKE operational overhead|Time sink|Offer Cloud Run default; GKE only when workload justifies|
|Evolving A2A spec|Integration churn|Wrapper adaptor layer; monitor upstream releases|

---

## 8  Appendix  

**Glossary**  
* **A2A Protocol** – open standard for agent-to-agent messaging and capability discovery (Google, 2025) :contentReference[oaicite:2]{index=2}  
* **Genkit Go** – open-source Go toolkit for building AI-powered services; includes agent primitives and observability :contentReference[oaicite:3]{index=3}  
* **Temporal** – durable workflow orchestration platform with Go SDK; offers local dev service and managed Temporal Cloud :contentReference[oaicite:4]{index=4}  

End of document.
