# DMN vs JDM — Research Findings

## 1. What they are

**DMN (Decision Model and Notation)**
- OMG standard (same body as BPMN/UML), currently DMN 1.6 (April 2026)
- XML-based interchange format
- Uses FEEL (Friendly Enough Expression Language) for expressions
- Decision Requirement Diagrams (DRDs) show how decisions/inputs relate
- Backed by Camunda, Drools/KIE, IBM ODM, Trisotech, and others

**JDM (JSON Decision Model)**
- JSON-based format used primarily by GoRules' zen-engine (open source, Rust core)
- Lightweight, git-friendly, easy to embed
- Not an OMG/ISO standard — de facto format tied to the GoRules ecosystem
- Simple, JS-like expression syntax rather than FEEL

### Practical differences that matter for choosing

| | DMN | JDM |
|---|---|---|
| Standardization | OMG standard, multi-vendor | Single-ecosystem (GoRules/zen-engine) |
| Format | XML | JSON |
| Expression language | FEEL | Simple expression syntax (JS-like) |
| Tooling maturity | Mature, enterprise-grade (Camunda, Drools, ODM) | Newer, lighter, growing fast |
| Embeddability | Heavier runtime | Very lightweight, easy to embed (Rust core, WASM-capable) |
| Best fit | Enterprise BPM stacks, regulated industries needing portability/audit trail | Developer-first apps wanting a fast, git-friendly, embeddable rules engine |

**Choosing between them in practice:** do you need vendor-neutral portability and business-analyst-facing tooling (→ DMN), or do you want something lightweight you can version-control and embed directly in a service (→ JDM/zen-engine)?

## 2. Runtime deployment model

| | Embedded | Separate service |
|---|---|---|
| **JDM (zen-engine)** | Native default — official Go binding (`gorules/zen-go`, cgo over Rust) and Python binding (`zen-engine` via PyO3). Loads JSON rule graph in-process, no network hop. | Optional — GoRules also offers a hosted/self-hosted engine + visual editor for centralized management. |
| **DMN** | Possible via Drools/KIE embedded as a Java library, or Camunda's standalone DMN engine JAR — but this still requires a JVM. | Default in practice for enterprise engines (Camunda, IBM ODM) — centralized governance, business-user authoring, non-developers can publish rule changes. |

**Key constraint for a Go/Python stack:** every mature DMN engine (Camunda, Drools/KIE, IBM ODM) is JVM-based. There is no production-grade native Go or Python DMN engine — DMN in a Go/Python shop almost always means calling a JVM-based decision service over REST/gRPC.

## 3. Enterprise backing comparison

**DMN**
- OMG-governed standard with a vendor-neutral conformance test suite (DMN TCK)
- Mid-2026 TCK conformance leaders: Goldman Sachs jDMN (100%), Trisotech (99.97%), IBM BAMOE and Apache KIE/Drools (99.91% each)
- Camunda's own conformance trails the pack (~84% DMN-Scala, ~81% Platform 7.21) — Camunda ≠ conformance leader
- Commercial ecosystem: IBM ODM, Red Hat Decision Manager, Pega, FICO Blaze Advisor, Progress Corticon, Camunda — full governance/audit/support tiers

**JDM**
- Single-vendor (GoRules), bootstrapped and customer-funded since 2023, no investors
- Used by Fortune 100 companies in financial services, insurance, logistics for compliance-sensitive rules
- Git-style branching, change requests, approval gates for audit trails; SOC 2 Type II
- No independent conformance suite, no second implementing vendor — GoRules explicitly wants tight control over the JDM spec/roadmap, not a multi-vendor standard

**Bottom line:** DMN wins on standards-body governance and multi-vendor conformance; JDM wins on operational fit for a Go/Python, JVM-free stack, backed by a single but credible vendor.

## 4. Persistence format

- XML is the OMG-mandated interchange/serialization format for DMN — required for cross-vendor portability and what the TCK tests against
- In practice, engines parse XML into their own internal runtime model; authoring tools (Camunda Modeler, Trisotech) only serialize to XML on save/export
- Some tools offer alternate authoring representations (e.g., DMN-Scala DSL) that compile to/from standard XML
- Known pain point: XML diffs poorly in git compared to JSON — part of JDM's appeal for dev-centric workflows

## 5. Getting started with DMN + Docker + Go

Two viable paths using the Drools/KIE engine (Apache-licensed, 99.91% TCK conformance):

**A. Custom Quarkus/Kogito service (stable, production-shaped)**
1. Author the `.dmn` file (DMN.new, Camunda Modeler, or VS Code DMN extension)
2. Scaffold a Quarkus project with the `kogito-quarkus-decisions` extension
3. Drop the `.dmn` file into `src/main/resources/` — Kogito auto-generates a matching REST endpoint
4. Verify locally via `mvn quarkus:dev` and the auto-generated Swagger UI
5. Build with `mvn clean package`, then `docker build -f src/main/docker/Dockerfile.jvm -t dmn-service .`
6. Run: `docker run -p 8080:8080 dmn-service` — REST endpoint live at `http://localhost:8080/<DecisionName>`
7. Call from Go with a plain `net/http` POST of JSON input, decode JSON output — no DMN-specific Go library needed
8. Add as a service in `docker-compose.yml` alongside the Go app

**B. Apache KIE Kogito JIT Runner (faster prototyping)**
- Prebuilt image: `docker.io/apache/incubator-kie-kogito-jit-runner:latest`
- Lets you submit a `.dmn` model at runtime over HTTP and evaluate it on the fly — no Maven project or custom build step
- Good for prototyping or workflows where the DMN model changes without wanting to rebuild/redeploy
- Trade-off: less of a stable, typed contract than a purpose-built Quarkus service; exact request/response API shape should be verified against current KIE docs before building against it

**Alternative (not recommended as default for this stack):** Camunda 8 (`camunda/zeebe` image) — heavier (full workflow engine, not just decisions) and lower DMN conformance (~81–84%) than Drools/KIE.

## 6. Overall recommendation for the Go/Python stack

- **If avoiding a JVM dependency is the priority:** JDM/zen-engine is the pragmatic choice — native Go and Python bindings, JSON persistence, git-friendly, SOC 2, real Fortune 100 usage, no separate service required.
- **If vendor-neutral portability, FEEL, DRDs, or multi-vendor standards conformance are required** (e.g., regulated industry, procurement checklist): DMN via a JVM-based decision service (Drools/KIE-backed) is the stronger fit, accepting the network hop and extra infra.
