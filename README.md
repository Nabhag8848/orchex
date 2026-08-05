# Orchex

**A durable workflow execution engine**

Build a graph. Publish an immutable version. Run it reliably. Resume exactly where it failed.

Design decisions, schemas, benches, and learning notes live in [docs/](./docs/).

## Design docs

| Path                                               | What it is                                                             |
| -------------------------------------------------- | ---------------------------------------------------------------------- |
| [docs/README.md](./docs/README.md)                 | Full design narrative (requirements → HLD → API → schema → deep dives) |
| [docs/orchex.excalidraw](./docs/orchex.excalidraw) | Architecture boards (source of truth for diagrams)                     |
| [docs/schema.dbml](./docs/schema.dbml)             | PostgreSQL OLTP model                                                  |
| [docs/bench/postgres](./docs/bench/postgres)       | OLTP capacity harness                                                  |
| [docs/node-type-schemas](./docs/node-type-schemas) | JSON Schema contracts for node types                                   |
| [docs/data-structure](./docs/data-structure)       | Graph experiments behind the DAG deep dive                             |

## License

[MIT](./LICENSE)
