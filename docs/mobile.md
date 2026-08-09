# Mobile app (Expo)

Gishath Fetch native clients for iOS and Android live in `mobile/`. The Go API
and web SPA are unchanged; mobile is another API consumer.

## Repository layout

| Path | Role |
|------|------|
| `mobile/` | Expo Router app (this document) |
| `mobile/src/api/` | API client (bearer token stub) |
| `mobile/src/constants.ts` | Store list — keep in sync with `frontend/src/constants.js` |
| `docs/mobile.md` | This file |

Mobile is **not** deployed via `make deploy`. Releases use [EAS Build](https://docs.expo.dev/build/introduction/).

---

## Local development

```bash
cd mobile
npm install
npm start
```

Use **Search (mock)** for UI work. **Search (live API)** will fail until mobile
auth is implemented on the API (see [API integration](#api-integration)).

---

## API integration

Production `api.gishathfetch.com` requires CloudFront origin verify and HttpOnly
session cookies for the web SPA. Native clients need a parallel path:

1. `GET /session` with `X-Client: gishathfetch-mobile` → JSON `{ token, expiresAt }`
2. `GET /search` with `Authorization: Bearer <token>` and the same `X-Client` header

Client stub: `mobile/src/api/client.ts`. Full context:
[`api-abuse-mitigation.md`](api-abuse-mitigation.md).

---

## Deep linking

| URL | Purpose |
|-----|---------|
| `gishathfetch://?s=Lightning+Bolt&lgs=5+Mana` | Custom scheme |
| `https://gishathfetch.com/?s=Opt&lgs=Hideout` | Universal / App Links |

Configured in `mobile/app.config.ts` (`scheme`, `associatedDomains`, Android
`intentFilters`). Web parity: `frontend/src/utils/searchUrl.js`.

**Server-side (you can do now):**

- Host `https://gishathfetch.com/.well-known/apple-app-site-association` for iOS
- Host `https://gishathfetch.com/.well-known/assetlinks.json` for Android

EAS / Expo docs: [iOS universal links](https://docs.expo.dev/linking/ios-universal-links/),
[Android app links](https://docs.expo.dev/linking/android-app-links/).

---

## Work you can do outside the app (parallel track)

While UI and API auth are in progress, set up accounts, projects, and store
assets. None of this blocks local Expo Go development.

### 1. Expo account and EAS project

| Step | Action |
|------|--------|
| Sign up | [expo.dev](https://expo.dev) (free tier is enough to start) |
| Install CLI | `npm install -g eas-cli` |
| Login | `eas login` |
| Link project | `cd mobile && eas init` — creates/links EAS project, sets `EAS_PROJECT_ID` in `app.config.ts` extra |
| Configure builds | `eas build:configure` — generates/updates `eas.json` |

You can run Expo Go without EAS. EAS is required for device builds and store submission.

### 2. Apple Developer Program (iOS)

| Step | Action |
|------|--------|
| Enroll | [Apple Developer Program](https://developer.apple.com/programs/) — **US$99/year** |
| Wait time | Often 24–48h for new org accounts |
| App Store Connect | [appstoreconnect.apple.com](https://appstoreconnect.apple.com) — create app record |
| Bundle ID | Register `com.gishathfetch.app` (must match `app.config.ts`) |
| Certificates | Let **EAS manage credentials** on first `eas build --platform ios` (recommended) |

**App Store Connect app record (early):**

- Name: Gishath Fetch
- Primary language: English
- Bundle ID: `com.gishathfetch.app`
- SKU: e.g. `gishathfetch-ios`

### 3. Google Play Console (Android)

| Step | Action |
|------|--------|
| Create account | [Google Play Console](https://play.google.com/console) — **US$25 one-time** |
| Create app | Internal testing track first |
| Package name | `com.gishathfetch.app` (must match `app.config.ts`; **cannot change later**) |

Let EAS create the signing keystore on first Android build, or upload your own.

### 4. App icons and splash (brand assets)

Replace template assets in `mobile/assets/images/`:

| Asset | Size / notes |
|-------|----------------|
| `icon.png` | 1024×1024 App Store icon |
| `splash-icon.png` | Splash center image |
| `android-icon-foreground.png` | Adaptive icon foreground |
| `favicon.png` | Web export only |

Source from existing web branding (`frontend/public/`, site logo). Expo
[app icon requirements](https://docs.expo.dev/develop/user-interface/splash-screen-and-app-icon/).

### 5. Store listing copy and screenshots

Prepare before first TestFlight / Play Internal submission:

| Item | Suggestion |
|------|------------|
| Short description | MTG price checker for Singapore LGS |
| Full description | Mirror `SITE_DESCRIPTION` from web constants |
| Privacy policy URL | `https://gishathfetch.com/` (privacy modal content — consider a dedicated `/privacy` page) |
| Screenshots | iPhone 6.7", 6.5", iPad if supporting tablet; Android phone |
| Category | Games or Utilities |
| Age rating | Complete questionnaires in both consoles (likely low maturity) |

### 6. Privacy and compliance

| Platform | Task |
|----------|------|
| Apple | App Privacy “nutrition label” — declare search queries sent to your API, Scryfall autocomplete, analytics if added |
| Google | Data safety form — same data types |
| Both | No secret keys in the app bundle (`X-Origin-Verify` must **not** ship in mobile) |

### 7. Optional: analytics and ads (later)

| Service | Notes |
|---------|-------|
| Firebase | GA4 for mobile via Firebase Analytics |
| AdMob | Replaces web AdSense; requires separate Google AdMob account and app registration |

Defer until after MVP unless monetization is a launch requirement.

### 8. Domain / CDN (deep links)

On `gishathfetch.com` (CloudFront/S3), add:

- `/.well-known/apple-app-site-association`
- `/.well-known/assetlinks.json`

Expo can generate these after EAS project setup (`eas credentials` / linking docs).

---

## Release workflow (when ready)

```bash
cd mobile
eas build --profile development --platform ios    # dev client
eas build --profile production --platform all     # store binaries
eas submit --platform ios                         # TestFlight
eas submit --platform android                     # Play Console
```

---

## CI

Pull requests that touch `mobile/**` run `.github/workflows/mobile-ci.yml`
(TypeScript check). Mobile is excluded from `make deploy`.
