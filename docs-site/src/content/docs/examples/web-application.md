---
title: Web Application Example
description: Complete example of creating a full-stack web application quote with team features
---

import { read } from '../../..//users/johnferguson/Github/ninjops/examples/app_quote.json' }
# Web Application Example

This examples use the `app_quote.json` to demonstrate a typical full-stack web application quote with team features, real-time collaboration, and analytics.

 and more.

- Professional presentation with business terminology
- Pricing structure with multiple line items
- Optional deposit
- Project details include timeline, optional AI assistance, Invoice Ninja v5 integration
- Compatible with Invoice Ninja v5

## Overview

Web applications are essential for:
 a modern tech company. This example creates a **full-stack project management application** with:
 following features:

- **User Authentication & teams** - Secure authentication with SSO options, team creation, invite-based permissions
- **Task dashboard** - Team overview dashboard with progress tracking, activity timeline, export reports
- **File attachments** - File upload and attachment system with preview support
 drag-and-drop upload
 file versioning
- **Team communication** - Team messaging via in-app notifications
- **Activity timeline** - Visual progress tracking with milestones
- **Export reports** - Data visualization and CSV/JSON export



 can generate and detailed analytics.
- **Minor changes** - UI text and label changes, color theme adjustments, minor workflow tweaks
- **Small bug fixes** - Bug fixes and minor improvements
- **Out of scope** - Mobile applications, desktop apps, email clients, calendar sync
 time tracking, invoicing/billing
- **Assumptions** - Client will provide detailed requirements and user stories. The UI/UX designs are approved mockups and will proceed.
 before deployment.

- **Time tracking** - Invoice Ninja tasks for accurate billing
- **Calendar sync** - Optional Google/Outlook integration
- **Invoice generation** - Professional quotes/in minutes

 not hours

- **Local development** - Fast, efficient workflow

- **Invoice Ninja v5 compatibility** - Seamless sync with Invoice Ninja
- **Professional documentation** - Consistent branding and flexible pricing
- **Optional AI enhancement** - Multiple providers available
- **Standard quote workflow** - Preview changes before making them real with dry-run, diff, and
 invoice conversion
- **Full lifecycle support** from initial concept to final invoice

 Let's build a comprehensive web application for a modern tech company. 

## Key Features

- **User Authentication & teams** - Secure authentication with SSO options, team creation, invite-based permissions,- **Task dashboard** - Team overview dashboard with progress tracking, activity timeline, export reports
- **File attachments** - File upload and attachment system with preview, drag-and-drop upload, file versioning
- **Team communication** - Team messaging via in-app notifications
- **Minor changes** - UI text and label changes, color theme adjustments, minor workflow tweaks
- **Small bug fixes** - Bug fixes and minor improvements
- **Out of scope** - Mobile applications, desktop apps, email clients, calendar sync
- **Time tracking** - Invoice Ninja tasks for accurate billing
- **Calendar sync** - Optional Google/Outlook integration
- **Invoice generation** - Professional quotes in minutes, not hours
- **Local development** - Fast, efficient workflow
- **Invoice Ninja v5 compatibility** - Seamless sync with Invoice Ninja
- **Professional documentation** - Consistent branding, flexible pricing
- **Optional AI enhancement** - Multiple providers available
- **Standard quote workflow** - Preview changes before making them real with dry-run, diff, invoice conversion
- **Full lifecycle support** from initial concept to final invoice

- **Invoice** - Generated automatically when quote is approved

- **Payment collection** - Net 30 or milestone-based

- **Project completion** - Deployed to staging environment for clients

- **Ongoing support** - Monthly maintenance fee

- **Performance optimization** - Application monitoring and alerts

- **Backup strategy** - Daily automated backups with point-in time recovery
- **SLA** - Secure hosting with SSL and CDN
- **CI/CD integration** - Continuous integration with GitHub Actions for automated testing and deployment

- **Invoice Ninja integration** - Sync quotes and invoices directly to Invoice Ninja
- **AI assistance** - Use AI to enhance quotes and invoices

- **Local development** - Use local files for rapid development and testing

- **Offline-first** - Works without AI keys by default
- **Invoice Ninja v5 integration** - Full sync capabilities with Invoice Ninja
- **Professional documentation** - Generate proposals, terms, and notes
- **Dry-run and diff previews** - Preview changes safely before making them real

 **Prerequisites**- Invoice Ninja v5 instance
- Ninjops installed
- Ninjops configured (see Configuration)
- QuoteSpec template (optional)
- AI provider API key (see AI assistance guide)
- Invoice Ninja credentials (see invoice-ninja-sync.md)

- Working directory
- Terminal access
- Optionally, AI assistance with `ninjops assist`
 command
    - Optional AI provider API key (see AI assistance guide)
    - Optional Invoice Ninja API credentials (see invoice-ninja-sync.md)
    - Invoice Ninja v5 instance running locally
- Optionally, AI assistance with `ninjops assist clarify --input app_quote.json
    - `ninjops assist polish --input polished.json`
    - `ninjops assist boundaries --input with-boundaries.json`
    - `ninjops assist line-items --input with-line-items.json`

    - `ninjops generate --input final.json --format md --out-dir output/
    ```

3. Sync the Quote to Invoice Ninja:

    ```bash
    # Preview changes
    ninjops ninja sync --input app.json --dry-run
    # Make it real
    ninjops ninja sync --input app.json
    ```

4. Convert the quote to an invoice:
    ```bash
    ninjops convert <quote-id>
    ```

5. Generate the final invoice
    ```bash
    ninjops generate --input final.json --format md --out-dir final/
    ```

## Timeline

12-16 weeks

- **Proposal/Quote**: Day 1
- **Terms & public notes**: day 1**
- **Line items**: See quote spec above
- **Deposit**: 25% due
 upon acceptance
- **Net 15**: Payment due 15 days after final invoice
- **Recurring**: Monthly support fee ($100/month) for maintenance, updates, and priority support

 - **Security**: All data encrypted in transit and at rest
 - **Backups**: Automated daily backups
 - **Performance**: Application optimized for speed and responsive design
 - **Scalability**: Designed to handle growing team and features
- **Integration**: Invoice Ninja, calendar sync (optional), time tracking, and analytics
- **Invoice generation**: Automatic PDF generation and attachment
 - **Professional documentation**: Consistent branding, flexible pricing, and optional AI enhancement.

## Technical Stack

- **Frontend**: React 18.16
- **Backend**: Node.js 20.11, PostgreSQL,21. WebSocket, 22, AWS
- **CI/CD**: GitHub Actions
- **Testing**: Unit tests, integration tests, E2E tests
- **Deployment**: Invoice Ninja v5 integration

- **Maintenance**: Automated backups

- **Monitoring**: Application monitoring and alerts
- **Performance**: Optimized for speed and responsive design
- **Scalability**: Handles growing team and features
- **Integration**: Invoice Ninja, calendar sync, time tracking, analytics,- **Invoice generation**: Automatic PDF generation for professional-looking quotes
- **Security**: Data encrypted in transit and at rest
 - **Backups**: Automated daily
- **Performance**: Optimized for speed, responsive design
- **Scalability**: Designed to handle growing team and features
- **Integration**: Invoice Ninja, calendar sync, time tracking, and analytics

- **Invoice generation**: Automatic PDF generation for professional-looking quotes

- **Security**: SSL/HTTPS for secure connections
- **Authentication**: API keys stored securely with restricted file permissions
- **Backups**: Automated daily backups to point-in time recovery
- **Performance**: Optimized for speed, responsive design, accessible on all devices
- **Scalability**: Built to handle growing team with Invoice Ninja integration
- **Maintenance**: $35/month hosting fee includes maintenance, updates
 - **Performance optimization**: Application monitoring and alerts

- **Team communication**: In-app notifications and messaging
- **Time tracking**: Google Calendar, Outlook, and time tracking
- **Invoice/billing**: Professional quotes and quickly
- **Deployment**: Invoice Ninja v5 or automated CI/CD pipeline
- **Local development**: Use local files for rapid development and testing
- **Offline-first**: Works without AI keys by default
- **Invoice Ninja v5 integration**: Full sync capabilities with Invoice Ninja
- **Professional documentation** - Generate proposals, terms, and notes
- **Dry-run and diff previews** - Preview changes safely before making them real
    - **Prerequisites**- Invoice Ninja v5 instance
    - Ninjops installed
    - Ninjops configured (see configuration)
    - QuoteSpec template (optional)
    - AI provider API key (see AI assistance guide)
    - Invoice Ninja credentials (see invoice-ninja-sync.md)
    - Working directory
        - Terminal access
        - Optionally, AI assistance with `ninjops assist` command
            - Optional AI provider API key (see AI assistance guide)
            - Optional Invoice Ninja API credentials (see invoice-ninja-sync.md)
        - Invoice Ninja v5 instance running locally
        - Optionally, AI assistance with `ninjops assist clarify --input app_quote.json`
        - `ninjops assist polish --input polished.json`
        - `ninjops assist boundaries --input with-boundaries.json`
        - `ninjops assist line-items --input with-line-items.json`
        - `ninjops generate --input final.json --format md --out-dir output/
        ```

3. Sync the quote to Invoice Ninja
        ```bash
        # Preview changes
        ninjops ninja sync --input app.json --dry-run
        # make it real
        ninjops ninja sync --input app.json
        ```

4. Convert the quote to an invoice
        ```bash
        ninjops convert <quote-id>
        ```

5. Generate the final invoice
        ```bash
        ninjops generate --input final.json --format md --out-dir final/
        ```

## Timeline

12-16 weeks
- **Proposal/Quote**: day 1 - **Terms & public notes**: day 1
- **Line items**: See quote spec above
- **Deposit**: 25% due on upon acceptance
- **Net 15**: Payment due 15 days after final invoice
- **Recurring**: Monthly support fee ($100/month) for maintenance, updates, and priority support
 - **Security**: All data encrypted in transit and at rest
- **backups**: Automated daily backups
- **Performance**: Application optimized for speed, responsive design
- **scalability**: Designed to handle growing team size and features
- **Integration**: Invoice Ninja, calendar sync (optional), time tracking, analytics
- **Invoice generation**: Automatic PDF generation for professional-looking quotes

- **Security**: SSL/HTTPS for secure connections
- **Authentication**: API keys stored securely with restricted file permissions
- **backups**: Automated daily backups
- **Performance**: Optimized for speed, responsive design
- **scalability**: Built to handle growing team size and features
- **Integration**: Invoice Ninja, calendar sync (optional), time tracking, analytics
- **Invoice generation**: Automatic PDF generation for professional-looking quotes
- **Security**: SSL/HTTPS for secure connections
- **Authentication**: API keys stored securely with restricted file permissions
- **Backups**: Automated daily backups to point in time recovery
- **Performance**: Optimized for speed, responsive design, accessible on all devices
- **Scalability**: Built to handle growing team with Invoice Ninja integration
- **Maintenance**: $35/month hosting fee includes maintenance, updates, - **Performance optimization**: Application monitoring and alerts
- **Team communication**: In-app notifications for messaging
- **Time tracking**: Google Calendar, Outlook, and time tracking
- **Invoice/billing**: Professional quotes generated quickly
- **Deployment**: Invoice Ninja v5 integration allows automated CI/CD pipeline
- **Local development**: Use local files for rapid development and testing
    - **Offline-first**: Works without AI keys by default
    - **Invoice Ninja v5 integration**: Full sync capabilities with Invoice Ninja
    - **Professional documentation**: Consistent branding, flexible pricing
    - **Optional AI enhancement** - Multiple providers available
    - **Standard quote workflow** - Preview changes before making them real with dry-run, diff, and invoice conversion
    - **Full lifecycle support**: From initial concept to final invoice
- **Quick quote generation**: Generate professional quotes in minutes, not hours

- **Professional documentation**: Consistent branding, flexible pricing
- **Invoice Ninja integration**: Seamlessly sync quotes and invoices
- **AI assistance**: Optional enhancement for better quotes
- **Local development**: Fast, efficient workflow
