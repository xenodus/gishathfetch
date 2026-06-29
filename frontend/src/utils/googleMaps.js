import { importLibrary, setOptions } from "@googlemaps/js-api-loader";

let loadPromise = null;

export function getGoogleMapsApiKey() {
  return import.meta.env.VITE_GOOGLE_MAPS_API_KEY?.trim() || "";
}

export function loadGoogleMaps() {
  const apiKey = getGoogleMapsApiKey();
  if (!apiKey) {
    return Promise.reject(
      new Error("Missing VITE_GOOGLE_MAPS_API_KEY environment variable"),
    );
  }

  if (!loadPromise) {
    setOptions({ key: apiKey, v: "weekly" });
    loadPromise = importLibrary("maps");
  }

  return loadPromise;
}
