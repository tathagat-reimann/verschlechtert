\# Architecture Document

> **Status:** [Draft / In review / Approved / Deprecated]  
> **Owner:** [Name or team]  
> **Last updated:** [YYYY-MM-DD]  
> **Reviewers:** [Names or teams]

## 1. Overview

### Purpose

This document defines the architecture and design principles used in this project.

### Goals

- The developers should follow this architecture, while editing, or adding new features.
- Any AI Agent must also follow the guidelines in this document.

## 2. Context and Requirements

### Business context

Inflation. Everything is getting expensive.
Also the raw materials and processes to manufatcture.
It is common now-a-days for many corporations to redesign a product, making it worse in quality to keep the price constant.

Few examples:

1. reduce quantity. Instead of 200 grams there is now only 150g in the packet, but the price is same.
2. quality of material is reduced, so the product does not work as good or needs renewal earlier than it used it (breaks early).

### Functional requirements

1. User can login using Google (and later other providers)
1. System supports DE and EN locale/language
1. User can create a report for a product that has become worse with time.
1. User can see existing reports, filter, search
1. User can comment on existing reports. Comment is editable and deleteable by the user who created it.
1. User can like existing report. Like can be taken away.
1. User can provide an alternative to the mentioned product in the report.
 

### Non-functional requirements

| Category | Requirement | Target |
|---|---|---|
| Availability | [Requirement] | [Target] |
| Performance | [Requirement] | [Target] |
| Scalability | [Requirement] | [Target] |
| Security | [Requirement] | [Target] |

## 3. Technologies used

Backend is build in Golang.
Database is Postgres.
Firebasestore will be used for storage.

Frontend is ts/js - vite
MUI components are used

## 4. Backend Overview

Backend is designed with **DDD** and **TDD**.

Two domains are **User** and **Report**

Each domain has it's own package/folder where all it's stuff goes.

User - comes from firebase and is persisted in the DB in the table appuser.

Report - created by user and maintains comments, likes, alternatives, images etc.

### Domain
First the domain is defined. This also covers all the validations possible.
The properties are kept private using lowercase, so that they are only editable via given functions.

### Persistence/repo
Each domain as a repo which is used to interact with the DB

### Service
Service sits on top of domain and exposes methods that operate on domain and also interact with the persistence layer.
It also interacts with firebasestore.

### Handlers
Handlers sit on the top and calls the service as required.

### Dataflow
Frontend (user interaction) -> calls an API endpoint -> Handler -> Service -> Domain/s -> Persistence

## 5. Security

Outsources to firebase

## 6. Reliability and Operations

- **Deployment model:** [Description]
- **Scaling strategy:** [Description]
- **Backup and recovery:** [RPO, RTO, and procedures]
- **Monitoring and alerting:** [Metrics, logs, dashboards, and alerts]
- **Disaster recovery:** [Plan]

## 7. Environments and Deployment

| Environment | Purpose | Infrastructure | Configuration |
|---|---|---|---|
| Development | quick local development | docker | .env files and ENV variables |
| Staging/test | test | TBD | TBD |
| Production | release | TBD | TBD |

## 8. Helpful commands

Testing with report: go test ./... -tags=integration -coverprofile=cover.out && go tool cover -html cover.out