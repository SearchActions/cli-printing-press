---
name: printing-press-catalog
description: Browse and install pre-built Go CLIs for popular APIs from the catalog
version: 0.4.0
allowed-tools:
  - Bash
  - Read
  - Write
  - Glob
  - Grep
  - WebFetch
  - AskUserQuestion
---

# /printing-press-catalog

Browse and install pre-built Go CLIs for popular APIs.

## Quick Start

```
/printing-press-catalog
/printing-press-catalog install stripe
/printing-press-catalog search auth
```

## Prerequisites

- Go 1.21+ installed
- The printing-press repo at ~/cli-printing-press

## Workflows

### List Catalog (no arguments)

When invoked with no arguments, list all available CLIs grouped by category.

1. Read all YAML files in ~/cli-printing-press/catalog/ using Glob + Read
2. Parse each file's name, display_name, description, category fields
3. Group by category and display:

```
Available CLIs (12 entries):

Payments:
  stripe - Payment processing and financial infrastructure API
  square - Payment processing and commerce API

Auth:
  stytch - Authentication and user management API

Email:
  sendgrid - Email delivery and marketing API

Communication:
  discord - Chat and community platform API
  twilio - Communication APIs for SMS, voice, and messaging
  front - Customer communication platform API

Developer Tools:
  github - Software development platform API
  digitalocean - Cloud infrastructure and developer platform API

Project Management:
  asana - Work management and project tracking API

CRM:
  hubspot - CRM contacts API

Example:
  petstore - Canonical OpenAPI example

Install any CLI: /printing-press-catalog install <name>
```

### Install (install <name>)

When invoked with `install <name>`:

1. Read ~/cli-printing-press/catalog/<name>.yaml
2. If file doesn't exist, show error: "No catalog entry for '<name>'. Run /printing-press-catalog to see available CLIs."
3. Extract spec_url from the catalog entry
4. Show preview: "Installing <display_name> CLI from <spec_url>"
5. Build the printing-press binary if needed:
   ```bash
   cd ~/cli-printing-press && go build -o ./printing-press ./cmd/printing-press
   ```
6. Download the spec and generate:
   ```bash
   curl -sL -o /tmp/catalog-spec-$$.yaml "<spec_url>"
   cd ~/cli-printing-press && ./printing-press generate \
     --spec /tmp/catalog-spec-$$.yaml \
     --output ./shelf/<name>-cli \
     --validate
   ```
7. If all quality gates pass, present the result:
   ```
   Generated <name>-cli with X resources.

   Try it:
     cd shelf/<name>-cli
     go install ./cmd/<name>-cli
     <name>-cli --help
     <name>-cli doctor
   ```
8. If gates fail, show the error and suggest: "Try /printing-press <display_name> API for a custom generation with retry support."

### Search (search <query>)

When invoked with `search <query>`:

1. Read all YAML files in ~/cli-printing-press/catalog/
2. Search name, display_name, description, and category for the query (case-insensitive)
3. Display matching entries

## Limitations

- Large API specs (Stripe, Discord, GitHub) take 30-60 seconds to generate and compile
- Generated CLIs are truncated to 50 resources / 20 endpoints per resource
- Catalog entries point to external URLs that may change
