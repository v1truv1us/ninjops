---
title: Architecture Overview
description: Learn about Ninjops architecture and design decisions
---

# Architecture Overview

Ninjops is built on a **deterministic-first** approach with optional AI enhancement and full Invoice Ninja v5 integration.

## Core Philosophy

### Deterministic-First
Works without AI keys. Templates and validation function independently. This provides several key benefits:

- **No external dependencies** - Everything works offline
- **Fast execution** - No network latency
- **Predictable** - Same inputs always produce the same output

- **Safe sync pattern** - GET → PATCH → PUT prevents data loss

### Invoice Ninja Integration
- **Full sync capabilities** with Invoice Ninja v5
- **Reference tracking** via `ninjops:<uuid>` tags
- **Safe update strategy** - Only updates changed fields, preserving remote state

- **Comprehensive field mapping** between QuoteSpec and Invoice Ninja

## Design Decisions

See [Architecture Decisions](/ninjops/architecture/decisions/) for detailed information.

Also see source specifications in the repository:
- [PRD](https://github.com/v1truv1us/ninjops/blob/main/docs/PRD.md)
- [TECH_SPEC](https://github.com/v1truv1us/ninjops/blob/main/docs/TECH_SPEC.md)
- [API_CONTRACT](https://github.com/v1truv1us/ninjops/blob/main/docs/API_CONTRACT.md)
