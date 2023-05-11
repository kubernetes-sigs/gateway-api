# GEP-X: Declarative Policy

* Issue: TODO
* Status: Provisional

## Definitions

In this document we'll use `Policy` to refer to things that are specifically called policies
as well as other "MetaResources" that follow similar patterns.

## TLDR

This proposal is a follow-up to [GEP-713 Metaresources and Policy Attachment][713] to recommend
that we specifically remove the "attachment" part of "policy attachment" in favor of something
that is declarative at the affected resource level.

[713]:https://gateway-api.sigs.k8s.io/geps/gep-713/

## Goals

- Remove "attachment" from `Policy` resources and related documentation.
- Retain `Policy` resource structure other than "attachment" semantics.
- Provide new semantics to incorporate `Policy` resources at the level of the `Resource` that
  will be affected.

## Introduction

TODO

## API

TODO: future iteration
