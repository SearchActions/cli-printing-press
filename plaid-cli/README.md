# plaid-cli

The Plaid REST API. Please see https://plaid.com/docs/api for more details.

## Install

### Homebrew

```
brew install USER/tap/plaid-cli
```

### Go

```
go install github.com/USER/plaid-cli/cmd/plaid-cli@latest
```

### Binary

Download from [Releases](https://github.com/USER/plaid-cli/releases).

## Quick Start

```bash
# 1. Set your API credentials
export PLAID_CLIENT_ID="your-key-here"

# 2. Verify everything works
plaid-cli doctor

# 3. Start using it
plaid-cli accounts --help
```

## Usage

<!-- HELP_OUTPUT -->

## Commands

### accounts

Manage accounts

- **`plaid-cli accounts balance-get`** - Retrieve real-time balance data
- **`plaid-cli accounts get`** - Retrieve accounts

### application

Manage application

- **`plaid-cli application get`** - Retrieve information about a Plaid application

### asset-report

Manage asset report

- **`plaid-cli asset-report audit-copy-create`** - Create Asset Report Audit Copy
- **`plaid-cli asset-report audit-copy-get`** - Retrieve an Asset Report Audit Copy
- **`plaid-cli asset-report audit-copy-pdf-get`** - Retrieve a PDF Asset Report Audit Copy
- **`plaid-cli asset-report audit-copy-remove`** - Remove Asset Report Audit Copy
- **`plaid-cli asset-report create`** - Create an Asset Report
- **`plaid-cli asset-report filter`** - Filter Asset Report
- **`plaid-cli asset-report get`** - Retrieve an Asset Report
- **`plaid-cli asset-report pdf-get`** - Retrieve a PDF Asset Report
- **`plaid-cli asset-report refresh`** - Refresh an Asset Report
- **`plaid-cli asset-report remove`** - Delete an Asset Report

### auth

Manage auth

- **`plaid-cli auth get`** - Retrieve auth data
- **`plaid-cli auth verify`** - Verify auth data

### bank-transfer

Manage bank transfer

- **`plaid-cli bank-transfer balance-get`** - Get balance of your Bank Transfer account
- **`plaid-cli bank-transfer cancel`** - Cancel a bank transfer
- **`plaid-cli bank-transfer create`** - Create a bank transfer
- **`plaid-cli bank-transfer event-list`** - List bank transfer events
- **`plaid-cli bank-transfer event-sync`** - Sync bank transfer events
- **`plaid-cli bank-transfer get`** - Retrieve a bank transfer
- **`plaid-cli bank-transfer list`** - List bank transfers
- **`plaid-cli bank-transfer migrate-account`** - Migrate account into Bank Transfers
- **`plaid-cli bank-transfer sweep-get`** - Retrieve a sweep
- **`plaid-cli bank-transfer sweep-list`** - List sweeps

### beacon

Manage beacon

- **`plaid-cli beacon account-risk-evaluate`** - Evaluate risk of a bank account
- **`plaid-cli beacon duplicate-get`** - Get a Beacon Duplicate
- **`plaid-cli beacon report-create`** - Create a Beacon Report
- **`plaid-cli beacon report-get`** - Get a Beacon Report
- **`plaid-cli beacon report-list`** - List Beacon Reports for a Beacon User
- **`plaid-cli beacon report-syndication-get`** - Get a Beacon Report Syndication
- **`plaid-cli beacon report-syndication-list`** - List Beacon Report Syndications for a Beacon User
- **`plaid-cli beacon user-account-insights-get`** - Get Account Insights for a Beacon User
- **`plaid-cli beacon user-create`** - Create a Beacon User
- **`plaid-cli beacon user-get`** - Get a Beacon User
- **`plaid-cli beacon user-history-list`** - List a Beacon User's history
- **`plaid-cli beacon user-review`** - Review a Beacon User
- **`plaid-cli beacon user-update`** - Update the identity data of a Beacon User

### beta

Manage beta

- **`plaid-cli beta credit-bank-employment-get`** - Retrieve information from the bank accounts used for employment verification
- **`plaid-cli beta ewa-report-v1-get`** - Get EWA Score Report
- **`plaid-cli beta partner-customer-v1-create`** - Creates a new end customer for a Plaid reseller.
- **`plaid-cli beta partner-customer-v1-enable`** - Enables a Plaid reseller's end customer in the Production environment.
- **`plaid-cli beta partner-customer-v1-get`** - Retrieves the details of a Plaid reseller's end customer.
- **`plaid-cli beta partner-customer-v1-update`** - Updates an existing end customer.
- **`plaid-cli beta transactions-enhance`** - enhance locally-held transaction data
- **`plaid-cli beta transactions-rules-create`** - Create transaction category rule
- **`plaid-cli beta transactions-rules-list`** - Return a list of rules created for the Item associated with the access token.
- **`plaid-cli beta transactions-rules-remove`** - Remove transaction rule
- **`plaid-cli beta transactions-user-insights-get`** - Obtain user insights based on transactions sent through /transactions/enrich

### business-verification

Manage business verification

- **`plaid-cli business-verification create`** - Create a business verification
- **`plaid-cli business-verification get`** - Get a business verification

### cashflow-report

Manage cashflow report

- **`plaid-cli cashflow-report get`** - Gets transaction data in `cashflow_report`
- **`plaid-cli cashflow-report insights-get`** - Gets insights data in Cashflow Report
- **`plaid-cli cashflow-report refresh`** - Refresh transaction data in `cashflow_report`
- **`plaid-cli cashflow-report transactions-get`** - Gets transaction data in cashflow_report

### categories

Manage categories

- **`plaid-cli categories get`** - (Deprecated) Get legacy categories

### consent

Manage consent

- **`plaid-cli consent events-get`** - List a historical log of item consent events

### consumer-report

Manage consumer report

- **`plaid-cli consumer-report pdf-get`** - Retrieve a PDF Reports

### cra

Manage cra

- **`plaid-cli cra check-report-base-report-get`** - Retrieve a Base Report
- **`plaid-cli cra check-report-cashflow-insights-get`** - Retrieve cash flow insights from your user's banking data
- **`plaid-cli cra check-report-create`** - Refresh or create a Consumer Report
- **`plaid-cli cra check-report-income-insights-get`** - Retrieve cash flow information from your user's banks
- **`plaid-cli cra check-report-lend-score-get`** - Retrieve the LendScore from your user's banking data
- **`plaid-cli cra check-report-network-insights-get`** - Retrieve network attributes for the user
- **`plaid-cli cra check-report-partner-insights-get`** - Retrieve cash flow insights from partners
- **`plaid-cli cra check-report-pdf-get`** - Retrieve Consumer Reports as a PDF
- **`plaid-cli cra check-report-verification-get`** - Retrieve various home lending reports for a user.
- **`plaid-cli cra check-report-verification-pdf-get`** - Retrieve Consumer Reports as a Verification PDF
- **`plaid-cli cra loans-applications-register`** - Register loan applications and decisions.
- **`plaid-cli cra loans-register`** - Register a list of loans to their applicants.
- **`plaid-cli cra loans-unregister`** - Unregister a list of loans.
- **`plaid-cli cra loans-update`** - Updates loan data.
- **`plaid-cli cra monitoring-insights-get`** - Retrieve a Monitoring Insights Report
- **`plaid-cli cra monitoring-insights-subscribe`** - Subscribe to Monitoring Insights
- **`plaid-cli cra monitoring-insights-unsubscribe`** - Unsubscribe from Monitoring Insights
- **`plaid-cli cra partner-insights-get`** - Retrieve cash flow insights from the bank accounts used for income verification

### credit

Manage credit

- **`plaid-cli credit asset-report-freddie-mac-get`** - Retrieve an Asset Report with Freddie Mac format. Only Freddie Mac can use this endpoint.
- **`plaid-cli credit audit-copy-token-create`** - Create Asset or Income Report Audit Copy Token
- **`plaid-cli credit audit-copy-token-update`** - Update an Audit Copy Token
- **`plaid-cli credit bank-income-get`** - Retrieve information from the bank accounts used for income verification
- **`plaid-cli credit bank-income-pdf-get`** - Retrieve information from the bank accounts used for income verification in PDF format
- **`plaid-cli credit bank-income-refresh`** - Refresh a user's bank income information
- **`plaid-cli credit bank-income-webhook-update`** - Subscribe and unsubscribe to proactive notifications for a user's income profile
- **`plaid-cli credit bank-statements-uploads-get`** - Retrieve data for a user's uploaded bank statements
- **`plaid-cli credit employment-get`** - Retrieve a summary of an individual's employment information
- **`plaid-cli credit freddie-mac-reports-get`** - Retrieve an Asset Report with Freddie Mac format (aka VOA - Verification Of Assets), and a Verification Of Employment (VOE) report if this one is available. Only Freddie Mac can use this endpoint.
- **`plaid-cli credit payroll-income-get`** - Retrieve a user's payroll information
- **`plaid-cli credit payroll-income-parsing-config-update`** - Update the parsing configuration for a document income verification
- **`plaid-cli credit payroll-income-precheck`** - Check income verification eligibility and optimize conversion
- **`plaid-cli credit payroll-income-refresh`** - Refresh a digital payroll income verification
- **`plaid-cli credit payroll-income-risk-signals-get`** - Retrieve fraud insights for a user's manually uploaded document(s).
- **`plaid-cli credit relay-create`** - Create a relay token to share an Asset Report with a partner client
- **`plaid-cli credit relay-get`** - Retrieve the reports associated with a relay token that was shared with you
- **`plaid-cli credit relay-pdf-get`** - Retrieve the pdf reports associated with a relay token that was shared with you (beta)
- **`plaid-cli credit relay-refresh`** - Refresh a report of a relay token
- **`plaid-cli credit relay-remove`** - Remove relay token
- **`plaid-cli credit report-audit-copy-remove`** - Remove an Audit Copy token
- **`plaid-cli credit sessions-get`** - Retrieve Link sessions for your user

### dashboard-user

Manage dashboard user

- **`plaid-cli dashboard-user get`** - Retrieve a dashboard user
- **`plaid-cli dashboard-user list`** - List dashboard users

### employers

Manage employers

- **`plaid-cli employers search`** - Search employer database

### employment

Manage employment

- **`plaid-cli employment verification-get`** - (Deprecated) Retrieve a summary of an individual's employment information

### fdx

Manage fdx

- **`plaid-cli fdx get-recipient`** - Get Recipient
- **`plaid-cli fdx get-recipients`** - Get Recipients
- **`plaid-cli fdx notifications`** - Webhook receiver for fdx notifications

### identity

Manage identity

- **`plaid-cli identity documents-uploads-get`** - Returns uploaded document identity
- **`plaid-cli identity get`** - Retrieve identity data
- **`plaid-cli identity match`** - Retrieve identity match score
- **`plaid-cli identity refresh`** - Refresh identity data

### identity-verification

Manage identity verification

- **`plaid-cli identity-verification autofill-create`** - Create autofill for an Identity Verification
- **`plaid-cli identity-verification create`** - Create a new Identity Verification
- **`plaid-cli identity-verification get`** - Retrieve Identity Verification
- **`plaid-cli identity-verification list`** - List Identity Verifications
- **`plaid-cli identity-verification retry`** - Retry an Identity Verification

### income

Manage income

- **`plaid-cli income verification-create`** - (Deprecated) Create an income verification instance
- **`plaid-cli income verification-documents-download`** - (Deprecated) Download the original documents used for income verification
- **`plaid-cli income verification-paystubs-get`** - (Deprecated) Retrieve information from the paystubs used for income verification
- **`plaid-cli income verification-precheck`** - (Deprecated) Check digital income verification eligibility and optimize conversion
- **`plaid-cli income verification-taxforms-get`** - (Deprecated) Retrieve information from the tax documents used for income verification

### institutions

Manage institutions

- **`plaid-cli institutions get`** - Get details of all supported institutions
- **`plaid-cli institutions get-by-id`** - Get details of an institution
- **`plaid-cli institutions search`** - Search institutions

### investments

Manage investments

- **`plaid-cli investments auth-get`** - Get data needed to authorize an investments transfer
- **`plaid-cli investments holdings-get`** - Get Investment holdings
- **`plaid-cli investments refresh`** - Refresh investment data
- **`plaid-cli investments transactions-get`** - Get investment transactions

### issues

Manage issues

- **`plaid-cli issues get`** - Get an Issue
- **`plaid-cli issues search`** - Search for an Issue
- **`plaid-cli issues subscribe`** - Subscribe to an Issue

### item

Manage item

- **`plaid-cli item access-token-invalidate`** - Invalidate access_token
- **`plaid-cli item activity-list`** - List a historical log of user consent events
- **`plaid-cli item application-list`** - List a user’s connected applications
- **`plaid-cli item application-scopes-update`** - Update the scopes of access for a particular application
- **`plaid-cli item application-unlink`** - Unlink a user’s connected application
- **`plaid-cli item create-public-token`** - Create public token
- **`plaid-cli item get`** - Retrieve an Item
- **`plaid-cli item import`** - Import Item
- **`plaid-cli item public-token-exchange`** - Exchange public token for an access token
- **`plaid-cli item remove`** - Remove an Item
- **`plaid-cli item webhook-update`** - Update Webhook URL

### liabilities

Manage liabilities

- **`plaid-cli liabilities get`** - Retrieve Liabilities data

### link

Manage link

- **`plaid-cli link oauth-correlation-id-exchange`** - Exchange the Link Correlation Id for a Link Token
- **`plaid-cli link token-create`** - Create Link Token
- **`plaid-cli link token-get`** - Get Link Token

### link-delivery

Manage link delivery

- **`plaid-cli link-delivery create`** - Create Hosted Link session
- **`plaid-cli link-delivery get`** - Get Hosted Link session

### network

Manage network

- **`plaid-cli network status-get`** - Check a user's Plaid Network status

### network-insights

Manage network insights

- **`plaid-cli network-insights report-get`** - Retrieve network insights for the provided `access_tokens`

### oauth

Manage oauth

- **`plaid-cli oauth introspect`** - Get metadata about an OAuth token
- **`plaid-cli oauth revoke`** - Revoke an OAuth token
- **`plaid-cli oauth token`** - Create or refresh an OAuth access token

### partner

Manage partner

- **`plaid-cli partner customer-create`** - Creates a new end customer for a Plaid reseller.
- **`plaid-cli partner customer-enable`** - Enables a Plaid reseller's end customer in the Production environment.
- **`plaid-cli partner customer-get`** - Returns a Plaid reseller's end customer.
- **`plaid-cli partner customer-oauth-institutions-get`** - Returns OAuth-institution registration information for a given end customer.
- **`plaid-cli partner customer-remove`** - Removes a Plaid reseller's end customer.

### payment-initiation

Manage payment initiation

- **`plaid-cli payment-initiation consent-create`** - Create payment consent
- **`plaid-cli payment-initiation consent-get`** - Get payment consent
- **`plaid-cli payment-initiation consent-payment-execute`** - Execute a single payment using consent
- **`plaid-cli payment-initiation consent-revoke`** - Revoke payment consent
- **`plaid-cli payment-initiation create-payment-token`** - Create payment token
- **`plaid-cli payment-initiation payment-create`** - Create a payment
- **`plaid-cli payment-initiation payment-get`** - Get payment details
- **`plaid-cli payment-initiation payment-list`** - List payments
- **`plaid-cli payment-initiation payment-reverse`** - Reverse an existing payment
- **`plaid-cli payment-initiation recipient-create`** - Create payment recipient
- **`plaid-cli payment-initiation recipient-get`** - Get payment recipient
- **`plaid-cli payment-initiation recipient-list`** - List payment recipients

### payment-profile

Manage payment profile

- **`plaid-cli payment-profile create`** - Create payment profile
- **`plaid-cli payment-profile get`** - Get payment profile
- **`plaid-cli payment-profile remove`** - Remove payment profile

### processor

Manage processor

- **`plaid-cli processor account-get`** - Retrieve the account associated with a processor token
- **`plaid-cli processor apex-token-create`** - Create Apex bank account token
- **`plaid-cli processor auth-get`** - Retrieve Auth data
- **`plaid-cli processor balance-get`** - Retrieve Balance data
- **`plaid-cli processor bank-transfer-create`** - Create a bank transfer as a processor
- **`plaid-cli processor identity-get`** - Retrieve Identity data
- **`plaid-cli processor identity-match`** - Retrieve identity match score
- **`plaid-cli processor investments-holdings-get`** - Retrieve Investment Holdings
- **`plaid-cli processor investments-transactions-get`** - Get investment transactions data
- **`plaid-cli processor liabilities-get`** - Retrieve Liabilities data
- **`plaid-cli processor signal-decision-report`** - Report whether you initiated an ACH transaction
- **`plaid-cli processor signal-evaluate`** - Evaluate a planned ACH transaction
- **`plaid-cli processor signal-prepare`** - Opt-in a processor token to Signal
- **`plaid-cli processor signal-return-report`** - Report a return for an ACH transaction
- **`plaid-cli processor stripe-bank-account-token-create`** - Create Stripe bank account token
- **`plaid-cli processor token-create`** - Create processor token
- **`plaid-cli processor token-permissions-get`** - Get a processor token's product permissions
- **`plaid-cli processor token-permissions-set`** - Control a processor's access to products
- **`plaid-cli processor token-webhook-update`** - Update a processor token's webhook URL
- **`plaid-cli processor transactions-get`** - Get transaction data
- **`plaid-cli processor transactions-recurring-get`** - Fetch recurring transaction streams
- **`plaid-cli processor transactions-refresh`** - Refresh transaction data
- **`plaid-cli processor transactions-sync`** - Get incremental transaction updates on a processor token

### profile

Manage profile

- **`plaid-cli profile network-status-get`** - Check a user's Plaid Network status

### protect

Manage protect

- **`plaid-cli protect compute`** - Compute Protect Trust Index Score
- **`plaid-cli protect event-get`** - Get information about a user event
- **`plaid-cli protect event-send`** - Send a new event to enrich user data
- **`plaid-cli protect report-create`** - Create a Protect report
- **`plaid-cli protect user-insights-get`** - Get Protect user insights

### sandbox

Manage sandbox

- **`plaid-cli sandbox bank-income-fire-webhook`** - Manually fire a bank income webhook in sandbox
- **`plaid-cli sandbox bank-transfer-fire-webhook`** - Manually fire a Bank Transfer webhook
- **`plaid-cli sandbox bank-transfer-simulate`** - Simulate a bank transfer event in Sandbox
- **`plaid-cli sandbox cra-cashflow-updates-update`** - Trigger an update for Cash Flow Updates
- **`plaid-cli sandbox income-fire-webhook`** - Manually fire an Income webhook
- **`plaid-cli sandbox item-fire-webhook`** - Fire a test webhook
- **`plaid-cli sandbox item-reset-login`** - Force a Sandbox Item into an error state
- **`plaid-cli sandbox item-set-verification-status`** - Set verification status for Sandbox account
- **`plaid-cli sandbox oauth-select-accounts`** - Save the selected accounts when connecting to the Platypus Oauth institution
- **`plaid-cli sandbox payment-profile-reset-login`** - Reset the login of a Payment Profile
- **`plaid-cli sandbox payment-simulate`** - Simulate a payment event in Sandbox
- **`plaid-cli sandbox processor-token-create`** - Create a test Item and processor token
- **`plaid-cli sandbox public-token-create`** - Create a test Item
- **`plaid-cli sandbox transactions-create`** - Create sandbox transactions
- **`plaid-cli sandbox transfer-fire-webhook`** - Manually fire a Transfer webhook
- **`plaid-cli sandbox transfer-ledger-deposit-simulate`** - Simulate a ledger deposit event in Sandbox
- **`plaid-cli sandbox transfer-ledger-simulate-available`** - Simulate converting pending balance to available balance
- **`plaid-cli sandbox transfer-ledger-withdraw-simulate`** - Simulate a ledger withdraw event in Sandbox
- **`plaid-cli sandbox transfer-refund-simulate`** - Simulate a refund event in Sandbox
- **`plaid-cli sandbox transfer-repayment-simulate`** - Trigger the creation of a repayment
- **`plaid-cli sandbox transfer-simulate`** - Simulate a transfer event in Sandbox
- **`plaid-cli sandbox transfer-sweep-simulate`** - Simulate creating a sweep
- **`plaid-cli sandbox transfer-test-clock-advance`** - Advance a test clock
- **`plaid-cli sandbox transfer-test-clock-create`** - Create a test clock
- **`plaid-cli sandbox transfer-test-clock-get`** - Get a test clock
- **`plaid-cli sandbox transfer-test-clock-list`** - List test clocks
- **`plaid-cli sandbox user-reset-login`** - Force item(s) for a Sandbox User into an error state

### session

Manage session

- **`plaid-cli session token-create`** - Create a Link token for Layer

### signal

Manage signal

- **`plaid-cli signal decision-report`** - Report whether you initiated an ACH transaction
- **`plaid-cli signal evaluate`** - Evaluate a planned ACH transaction
- **`plaid-cli signal prepare`** - Opt-in an Item to Signal Transaction Scores
- **`plaid-cli signal return-report`** - Report a return for an ACH transaction
- **`plaid-cli signal schedule`** - Schedule a planned ACH transaction

### statements

Manage statements

- **`plaid-cli statements download`** - Retrieve a single statement.
- **`plaid-cli statements list`** - Retrieve a list of all statements associated with an item.
- **`plaid-cli statements refresh`** - Refresh statements data.

### transactions

Manage transactions

- **`plaid-cli transactions enrich`** - Enrich locally-held transaction data
- **`plaid-cli transactions get`** - Get transaction data
- **`plaid-cli transactions recurring-get`** - Fetch recurring transaction streams
- **`plaid-cli transactions refresh`** - Refresh transaction data
- **`plaid-cli transactions sync`** - Get incremental transaction updates on an Item

### transfer

Manage transfer

- **`plaid-cli transfer authorization-cancel`** - Cancel a transfer authorization
- **`plaid-cli transfer authorization-create`** - Create a transfer authorization
- **`plaid-cli transfer balance-get`** - (Deprecated) Retrieve a balance held with Plaid
- **`plaid-cli transfer cancel`** - Cancel a transfer
- **`plaid-cli transfer capabilities-get`** - Get RTP eligibility information of a transfer
- **`plaid-cli transfer configuration-get`** - Get transfer product configuration
- **`plaid-cli transfer create`** - Create a transfer
- **`plaid-cli transfer diligence-document-upload`** - Upload transfer diligence document on behalf of the originator
- **`plaid-cli transfer diligence-submit`** - Submit transfer diligence on behalf of the originator
- **`plaid-cli transfer event-list`** - List transfer events
- **`plaid-cli transfer event-sync`** - Sync transfer events
- **`plaid-cli transfer get`** - Retrieve a transfer
- **`plaid-cli transfer intent-create`** - Create a transfer intent object to invoke the Transfer UI
- **`plaid-cli transfer intent-get`** - Retrieve more information about a transfer intent
- **`plaid-cli transfer ledger-deposit`** - Deposit funds into a Plaid Ledger balance
- **`plaid-cli transfer ledger-distribute`** - Move available balance between ledgers
- **`plaid-cli transfer ledger-event-list`** - List transfer ledger events
- **`plaid-cli transfer ledger-get`** - Retrieve Plaid Ledger balance
- **`plaid-cli transfer ledger-withdraw`** - Withdraw funds from a Plaid Ledger balance
- **`plaid-cli transfer list`** - List transfers
- **`plaid-cli transfer metrics-get`** - Get transfer product usage metrics
- **`plaid-cli transfer migrate-account`** - Migrate account into Transfers
- **`plaid-cli transfer originator-create`** - Create a new originator
- **`plaid-cli transfer originator-funding-account-create`** - Create a new funding account for an originator
- **`plaid-cli transfer originator-funding-account-update`** - Update the funding account associated with the originator
- **`plaid-cli transfer originator-get`** - Get status of an originator's onboarding
- **`plaid-cli transfer originator-list`** - Get status of all originators' onboarding
- **`plaid-cli transfer platform-originator-create`** - Create an originator for Transfer for Platforms customers
- **`plaid-cli transfer platform-person-create`** - Create a person associated with an originator
- **`plaid-cli transfer platform-requirement-submit`** - Submit additional onboarding information on behalf of an originator
- **`plaid-cli transfer questionnaire-create`** - Generate a Plaid-hosted onboarding UI URL.
- **`plaid-cli transfer recurring-cancel`** - Cancel a recurring transfer.
- **`plaid-cli transfer recurring-create`** - Create a recurring transfer
- **`plaid-cli transfer recurring-get`** - Retrieve a recurring transfer
- **`plaid-cli transfer recurring-list`** - List recurring transfers
- **`plaid-cli transfer refund-cancel`** - Cancel a refund
- **`plaid-cli transfer refund-create`** - Create a refund
- **`plaid-cli transfer refund-get`** - Retrieve a refund
- **`plaid-cli transfer repayment-list`** - Lists historical repayments
- **`plaid-cli transfer repayment-return-list`** - List the returns included in a repayment
- **`plaid-cli transfer sweep-get`** - Retrieve a sweep
- **`plaid-cli transfer sweep-list`** - List sweeps

### user

Manage user

- **`plaid-cli user create`** - Create user
- **`plaid-cli user financial-data-refresh`** - Refresh user items for Financial-Insights bundle
- **`plaid-cli user get`** - Retrieve user identity and information
- **`plaid-cli user identity-remove`** - Remove user identity data
- **`plaid-cli user items-associate`** - Associate Items to a User
- **`plaid-cli user items-get`** - Get Items associated with a User
- **`plaid-cli user items-remove`** - Remove Items from a User
- **`plaid-cli user products-terminate`** - Terminate user-based products
- **`plaid-cli user remove`** - Remove user
- **`plaid-cli user third-party-token-create`** - Create a third-party user token
- **`plaid-cli user third-party-token-remove`** - Remove a third-party user token
- **`plaid-cli user transactions-refresh`** - Refresh user items for Transactions bundle
- **`plaid-cli user update`** - Update user information

### user-account

Manage user account

- **`plaid-cli user-account session-event-send`** - Send User Account Session Event
- **`plaid-cli user-account session-get`** - Retrieve User Account

### wallet

Manage wallet

- **`plaid-cli wallet create`** - Create an e-wallet
- **`plaid-cli wallet get`** - Fetch an e-wallet
- **`plaid-cli wallet list`** - Fetch a list of e-wallets
- **`plaid-cli wallet transaction-execute`** - Execute a transaction using an e-wallet
- **`plaid-cli wallet transaction-get`** - Fetch an e-wallet transaction
- **`plaid-cli wallet transaction-list`** - List e-wallet transactions

### watchlist-screening

Manage watchlist screening

- **`plaid-cli watchlist-screening entity-create`** - Create a watchlist screening for an entity
- **`plaid-cli watchlist-screening entity-get`** - Get an entity screening
- **`plaid-cli watchlist-screening entity-history-list`** - List history for entity watchlist screenings
- **`plaid-cli watchlist-screening entity-hit-list`** - List hits for entity watchlist screenings
- **`plaid-cli watchlist-screening entity-list`** - List entity watchlist screenings
- **`plaid-cli watchlist-screening entity-program-get`** - Get entity watchlist screening program
- **`plaid-cli watchlist-screening entity-program-list`** - List entity watchlist screening programs
- **`plaid-cli watchlist-screening entity-review-create`** - Create a review for an entity watchlist screening
- **`plaid-cli watchlist-screening entity-review-list`** - List reviews for entity watchlist screenings
- **`plaid-cli watchlist-screening entity-update`** - Update an entity screening
- **`plaid-cli watchlist-screening individual-create`** - Create a watchlist screening for a person
- **`plaid-cli watchlist-screening individual-get`** - Retrieve an individual watchlist screening
- **`plaid-cli watchlist-screening individual-history-list`** - List history for individual watchlist screenings
- **`plaid-cli watchlist-screening individual-hit-list`** - List hits for individual watchlist screening
- **`plaid-cli watchlist-screening individual-list`** - List Individual Watchlist Screenings
- **`plaid-cli watchlist-screening individual-program-get`** - Get individual watchlist screening program
- **`plaid-cli watchlist-screening individual-program-list`** - List individual watchlist screening programs
- **`plaid-cli watchlist-screening individual-review-create`** - Create a review for an individual watchlist screening
- **`plaid-cli watchlist-screening individual-review-list`** - List reviews for individual watchlist screenings
- **`plaid-cli watchlist-screening individual-update`** - Update individual watchlist screening

### webhook-verification-key

Manage webhook verification key

- **`plaid-cli webhook-verification-key get`** - Get webhook verification key


## Output Formats

```bash
# Human-readable table (default)
plaid-cli accounts list

# JSON for scripting and agents
plaid-cli accounts list --json

# Filter specific fields
plaid-cli accounts list --json --select id,name,status

# Plain tab-separated for piping
plaid-cli accounts list --plain

# Dry run (show request without sending)
plaid-cli accounts list --dry-run
```

## Agent Usage

This CLI is designed for AI agent consumption:

```bash
# All commands support --json for structured output
plaid-cli accounts list --json --select id,name

# --dry-run shows the exact API request without sending
plaid-cli accounts list --dry-run

# Non-interactive - never prompts, never pages
# Errors go to stderr with typed exit codes
```

Exit codes: `0` success, `2` usage error, `3` not found, `4` auth error, `5` API error, `7` rate limited, `10` config error.

## Health Check

```bash
plaid-cli doctor
```

<!-- DOCTOR_OUTPUT -->

## Configuration

Config file: `~/.config/plaid-cli/config.toml`

Environment variables:
- `PLAID_CLIENT_ID`

## Troubleshooting

**Authentication errors (exit code 4)**
- Run `plaid-cli doctor` to check credentials
- Verify the environment variable is set: `echo $PLAID_CLIENT_ID`

**Not found errors (exit code 3)**
- Check the resource ID is correct
- Run the `list` command to see available items

**Rate limit errors (exit code 7)**
- The CLI auto-retries with exponential backoff
- If persistent, wait a few minutes and try again

---

Generated by [CLI Printing Press](https://github.com/mvanhorn/cli-printing-press)
