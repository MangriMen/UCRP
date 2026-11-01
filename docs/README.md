# Documentation

Project documentation for UCRP, a unified recommendation platform for shopping centers. The documents are written in Russian; READMEs are in English.

## Where things live

| Location               | What belongs there                                                                          | Lifecycle                   |
| ---------------------- | ------------------------------------------------------------------------------------------- | --------------------------- |
| `../ARCHITECTURE.md`   | Stable technical rules: layers, dependency direction, module boundaries                     | Changes rarely              |
| `requirements/`        | What the system must do: customer input, assumptions, use cases, open questions             | Updated when scope changes  |
| `design/`              | How it is designed: data model, diagrams, CRUD matrix, traceability, registry               | Updated together with code  |
| `../packages/openapi/` | API contract in YAML, the **source of truth** for paths, schemas and roles                  | Updated with every endpoint |
| `proposals/`           | Ideas not accepted yet. Move to `design/` once agreed, delete if rejected                   | Short-lived                 |
| `e2e/`                 | E2E test analysis and plan. Kept as is until its fate is decided                            | Not maintained              |
| `../archive/`          | Finished planning artifacts kept for history. Frozen; live docs may cite it only as history | Never updated               |

Rule of thumb: a document that describes the system as it is or must be goes to `requirements/` or `design/`. A document about how the work was planned or analyzed, once that work is done, goes to `archive/`.

## Requirements

| File                                     | Content                                                      |
| ---------------------------------------- | ------------------------------------------------------------ |
| `requirements/case.txt`                  | Original case statement                                      |
| `requirements/customer-faq.md`           | Customer answers to 10 questions; wins over everything else  |
| `requirements/01-summary-assumptions.md` | System summary, assumptions A-1…A-23, scope                  |
| `requirements/02-glossary.md`            | Roles R-1…R-6, entities E-1…E-22 and T-1…T-4, terms, metrics |
| `requirements/03-use-cases.md`           | Use case diagrams and the UC-1…UC-21 table                   |
| `requirements/04-user-stories.md`        | US-1…US-11 with Given/When/Then criteria                     |
| `requirements/00-open-questions.md`      | Questions Q-1…Q-41 and decisions D-1…D-39 with their status  |

## Design

| File                             | Content                                                                          |
| -------------------------------- | -------------------------------------------------------------------------------- |
| `design/05-er-data-model.md`     | ER diagrams, 3NF, extensibility, SQLite vs PostgreSQL, migrations, access tables |
| `design/06-object-diagrams.md`   | Data scenarios: phone merge, disputed merge and unmerge, segment export          |
| `design/07-crud-matrix.md`       | Role × entity matrix, permission layers, tenant isolation                        |
| `design/08-activity-diagrams.md` | Activity diagrams AD-1a…AD-5                                                     |
| `design/09-sequence-diagrams.md` | Sequence diagrams SD-1…SD-9                                                      |
| `design/10-class-diagram.md`     | Go interfaces and contracts CL-1…CL-6                                            |
| `design/11-component-diagram.md` | Components CMP-1…CMP-11, ports, scaling, change impact                           |
| `design/12-c4.md`                | C4 levels 1 to 4                                                                 |
| `design/13-dfd.md`               | Data flow diagrams, data stores D1…D7                                            |
| `design/14-traceability.md`      | US/UC → activity → sequence → ER → CRUD → component → API                        |
| `design/endpoint-registry.md`    | REST endpoints by role, queue job types                                          |

## Proposals

| File                                | Content                                            |
| ----------------------------------- | -------------------------------------------------- |
| `proposals/recommendation-model.md` | Proposed recommender model for guests and segments |

## Reading order

Quick start: `requirements/01` → `03` → `04`, then `design/05` → `09` → `12` → `14`. Decisions and open questions: `requirements/00-open-questions.md`. Before changing code: `../ARCHITECTURE.md`. A requirement change is checked against the impact matrix in `design/11-component-diagram.md` and the traceability table in `design/14-traceability.md`.
