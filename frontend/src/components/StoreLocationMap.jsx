import { useEffect, useRef, useState } from "react";
import { loadGoogleMaps } from "../utils/googleMaps";

const SINGAPORE_CENTER = { lat: 1.3521, lng: 103.8198 };
const DEFAULT_ZOOM = 11;

function escapeHtml(text) {
  return text
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function buildInfoWindowContent(shop) {
  const searchUrl = `/?src=${encodeURIComponent(shop.searchStore)}`;
  return `
    <div class="store-map-popup">
      <strong>${escapeHtml(shop.name)}</strong>
      <div class="mt-1 small">${escapeHtml(shop.address)}</div>
      <div class="mt-2">
        <a href="${escapeHtml(shop.website)}" target="_blank" rel="noreferrer">Website</a>
        &nbsp;|&nbsp;
        <a href="${searchUrl}">Search store</a>
      </div>
    </div>
  `;
}

const StoreLocationMap = ({ stores, activeStoreId, onMarkerClick }) => {
  const mapContainerRef = useRef(null);
  const mapRef = useRef(null);
  const markersRef = useRef({});
  const infoWindowRef = useRef(null);
  const onMarkerClickRef = useRef(onMarkerClick);
  const [loadError, setLoadError] = useState(null);

  useEffect(() => {
    onMarkerClickRef.current = onMarkerClick;
  }, [onMarkerClick]);

  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current) {
      return;
    }

    let isCancelled = false;

    loadGoogleMaps()
      .then(({ Map: GoogleMap }) => {
        if (isCancelled || !mapContainerRef.current) {
          return;
        }

        const map = new GoogleMap(mapContainerRef.current, {
          center: SINGAPORE_CENTER,
          zoom: DEFAULT_ZOOM,
          mapTypeControl: false,
          streetViewControl: false,
          fullscreenControl: true,
        });

        const infoWindow = new google.maps.InfoWindow();
        const markers = {};

        for (const shop of stores) {
          const marker = new google.maps.Marker({
            map,
            position: { lat: shop.lat, lng: shop.lng },
            title: shop.name,
          });

          marker.addListener("click", () => {
            infoWindow.setContent(buildInfoWindowContent(shop));
            infoWindow.open({ map, anchor: marker });
            onMarkerClickRef.current?.(shop.id);
          });

          markers[shop.id] = marker;
        }

        if (stores.length > 0) {
          const bounds = new google.maps.LatLngBounds();
          for (const shop of stores) {
            bounds.extend({ lat: shop.lat, lng: shop.lng });
          }
          map.fitBounds(bounds, 48);
        }

        mapRef.current = map;
        markersRef.current = markers;
        infoWindowRef.current = infoWindow;
        setLoadError(null);
      })
      .catch((error) => {
        if (!isCancelled) {
          setLoadError(error.message);
        }
      });

    return () => {
      isCancelled = true;
      infoWindowRef.current?.close();
      infoWindowRef.current = null;
      mapRef.current = null;
      markersRef.current = {};
    };
  }, [stores]);

  useEffect(() => {
    if (!activeStoreId || !mapRef.current || !infoWindowRef.current) {
      return;
    }

    const marker = markersRef.current[activeStoreId];
    const shop = stores.find((store) => store.id === activeStoreId);
    if (!marker || !shop) {
      return;
    }

    mapRef.current.panTo({ lat: shop.lat, lng: shop.lng });
    mapRef.current.setZoom(14);
    infoWindowRef.current.setContent(buildInfoWindowContent(shop));
    infoWindowRef.current.open({ map: mapRef.current, anchor: marker });
  }, [activeStoreId, stores]);

  if (loadError) {
    return (
      <div className="store-location-map border border-dark d-flex align-items-center justify-content-center text-muted p-3 text-center">
        Google Maps could not be loaded. Set{" "}
        <code className="mx-1">VITE_GOOGLE_MAPS_API_KEY</code> for the
        interactive map.
      </div>
    );
  }

  return (
    <div
      ref={mapContainerRef}
      className="store-location-map border border-dark"
      role="application"
      aria-label="Map of Singapore MTG store locations"
    />
  );
};

export default StoreLocationMap;
