import type { ExpoConfig } from "expo/config";

const apiBaseUrl =
  process.env.EXPO_PUBLIC_API_BASE_URL ?? "https://api.gishathfetch.com";
const siteBaseUrl =
  process.env.EXPO_PUBLIC_SITE_BASE_URL ?? "https://gishathfetch.com";

export default ({ config }: { config: ExpoConfig }): ExpoConfig => ({
  ...config,
  name: "Gishath Fetch",
  slug: "gishathfetch",
  version: "1.0.0",
  orientation: "portrait",
  icon: "./assets/images/icon.png",
  scheme: "gishathfetch",
  userInterfaceStyle: "automatic",
  ios: {
    supportsTablet: true,
    bundleIdentifier: "com.gishathfetch.app",
    associatedDomains: ["applinks:gishathfetch.com"],
  },
  android: {
    package: "com.gishathfetch.app",
    adaptiveIcon: {
      backgroundColor: "#E6F4FE",
      foregroundImage: "./assets/images/android-icon-foreground.png",
      backgroundImage: "./assets/images/android-icon-background.png",
      monochromeImage: "./assets/images/android-icon-monochrome.png",
    },
    predictiveBackGestureEnabled: false,
    intentFilters: [
      {
        action: "VIEW",
        autoVerify: true,
        data: [
          {
            scheme: "https",
            host: "gishathfetch.com",
            pathPrefix: "/",
          },
        ],
        category: ["BROWSABLE", "DEFAULT"],
      },
    ],
  },
  web: {
    bundler: "metro",
    output: "static",
    favicon: "./assets/images/favicon.png",
  },
  plugins: [
    "expo-router",
    [
      "expo-splash-screen",
      {
        image: "./assets/images/splash-icon.png",
        resizeMode: "contain",
        backgroundColor: "#ffffff",
      },
    ],
  ],
  experiments: {
    typedRoutes: true,
  },
  extra: {
    apiBaseUrl,
    siteBaseUrl,
    eas: {
      projectId: process.env.EAS_PROJECT_ID ?? "",
    },
  },
});
