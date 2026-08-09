# Gishath Fetch — Mobile (Expo)

React Native app for iOS and Android. Sibling to `frontend/` in the monorepo.

## Prerequisites

- Node 22
- [Expo Go](https://expo.dev/go) on a phone, or iOS Simulator / Android emulator
- macOS + Xcode for iOS simulator and App Store builds

## Quick start

```bash
cd mobile
npm install
npm start
```

Then press `i` (iOS), `a` (Android), or scan the QR code with Expo Go.

Optional env overrides: copy `.env.example` to `.env.local`.

## Scripts

| Command | Purpose |
|---------|---------|
| `npm start` | Expo dev server |
| `npm run ios` | Open iOS simulator |
| `npm run android` | Open Android emulator |
| `npm run lint` | TypeScript check (`tsc --noEmit`) |

## Project status

- **Search (mock)** — UI + Scryfall autocomplete; uses `src/data/mockSearchResults.ts`
- **Search (live API)** — stub in `src/api/client.ts`; requires backend mobile bearer-token auth
- **Saved / cart** — placeholder tab
- **Map** — store list with Maps / website links
- **FAQ** — placeholder; links to website

See [`docs/mobile.md`](../../docs/mobile.md) for store setup, EAS, and deep links.

## Bundle identifiers (locked before first EAS build)

| Platform | ID |
|----------|-----|
| iOS | `com.gishathfetch.app` |
| Android | `com.gishathfetch.app` |

Change these in `app.config.ts` only before your first production build.
