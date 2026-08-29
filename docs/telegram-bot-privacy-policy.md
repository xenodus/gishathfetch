# Gishath Fetch Telegram Bot — Privacy Policy

**Canonical URL:** [https://gishathfetch.com/telegram-bot-privacy.html](https://gishathfetch.com/telegram-bot-privacy.html)

**Last updated:** 29 August 2026

This policy describes how the **Gishath Fetch Telegram bot** (`@GishathFetchBot`) handles
information when you use it. It applies only to the bot. The main Gishath Fetch website
([gishathfetch.com](https://gishathfetch.com)) has a separate privacy policy (available
from the site footer).

## Summary

We do **not** collect or store personally identifiable information about Telegram users
(for example your Telegram user ID, username, display name, phone number, or chat ID).

The only user-provided content we retain for service operation and analytics is the
**card name or search text** you send with commands such as `/price` or `/ck`.

## What the bot does

The bot lets you look up Singapore MTG singles prices on Telegram:

| Command | Purpose |
|---------|---------|
| `/price <card name>` | Cheapest in-stock match across supported stores |
| `/ck <card name>` | Card Kingdom price from our database |
| `/help` | Usage instructions |

When you search, the bot calls the same Gishath Fetch search backend used by the website
and replies with prices and links back to [gishathfetch.com](https://gishathfetch.com).

## Information we collect

### Search terms (card names)

When you run `/price` or `/ck`, we process the **search text** you provide (for example
`Lightning Bolt`). That text may be:

- Used to run your search and return a result
- Recorded in aggregate analytics (for example search counts that can appear in the
  website **Trending** feature), in the same way as searches on the main site
- Written to short-lived operational logs (for example error diagnostics) on our hosting
  provider (Amazon Web Services)

We do **not** tie search terms to your Telegram account or identity in our systems.

### What we do not collect or store

We do **not** persistently store:

- Your Telegram user ID, username, or display name
- Your chat ID or group membership
- Your phone number or contact list
- Message content other than the search text you send for a price lookup
- Location data

Telegram identifiers are used **only in memory** for the duration of a request so the
bot can reply to the correct chat (including ForceReply prompts in group chats). They are
not written to our application database.

## Telegram and third parties

- **Telegram** delivers messages between you and the bot under
  [Telegram's Privacy Policy](https://telegram.org/privacy). We do not control Telegram's
  processing of your account or messages on their platform.
- **Store websites** linked in bot replies are operated by third parties with their own
  policies.
- **Hosting (AWS)** may generate standard infrastructure access logs (for example request
  timestamps and IP addresses at the network edge) as part of running the service. We do
  not use those logs to build profiles of individual Telegram users.

## How we use information

Search terms are used solely to:

1. Perform the price lookup you requested
2. Operate, secure, and improve Gishath Fetch (including aggregate usage statistics)
3. Diagnose failures when something goes wrong

We do not sell user data. We do not use Telegram data for advertising targeted at you as
an identifiable person.

## Data retention

Search terms used for analytics are kept in aggregate form according to our analytics
configuration (aligned with the main site's search analytics). Operational logs are
retained only as long as needed for troubleshooting and security, then rotated or deleted
according to our cloud provider's log retention settings.

We do not maintain a history of which Telegram users searched for which cards.

## Your choices

- **Stop using the bot:** block or stop messaging `@GishathFetchBot` at any time.
- **Use the website instead:** [gishathfetch.com](https://gishathfetch.com) offers the
  full multi-store search experience without Telegram.

## Children

The bot is intended for general audiences interested in trading card game prices. We do not
knowingly collect personal information from children.

## Changes

We may update this policy from time to time. The "Last updated" date at the top will change
when we do. Continued use of the bot after changes means you accept the updated policy.

## Contact

Questions about this policy or the bot:

**Email:** [contact@alvinyeoh.com](mailto:contact@alvinyeoh.com)
