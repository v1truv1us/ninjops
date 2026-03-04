---
title: Architecture Decisions
description: Key architectural decisions that guide Ninjops implementation
---

# Architecture Decisions

This page captures the high-level choices behind Ninjops behavior.

## Decision Highlights

## 1) Deterministic-first execution

Ninjops works without AI providers by default so core quote workflows remain reliable and reproducible.

## 2) Safe sync lifecycle

Sync operations follow a conservative read-compare-update pattern to reduce accidental remote data loss.

## 3) Structured QuoteSpec input

A typed QuoteSpec format keeps generation and validation consistent across CLI and API usage.

## 4) Optional AI augmentation

AI assistance is additive and does not block deterministic CLI operations.
