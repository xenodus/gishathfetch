import L from "leaflet";
import markerIcon from "leaflet/dist/images/marker-icon.png";
import markerIcon2x from "leaflet/dist/images/marker-icon-2x.png";
import markerShadow from "leaflet/dist/images/marker-shadow.png";
import { useEffect, useRef } from "react";
import "leaflet/dist/leaflet.css";

const defaultIcon = L.icon({
  iconUrl: markerIcon,
  iconRetinaUrl: markerIcon2x,
  shadowUrl: markerShadow,
  iconSize: [25, 41],
  iconAnchor: [12, 41],
  popupAnchor: [1, -34],
  shadowSize: [41, 41],
});

const SINGAPORE_CENTER = [1.3521, 103.8198];
const DEFAULT_ZOOM = 11;

function buildPopupContent(shop) {
  const searchUrl = `/?src=${encodeURIComponent(shop.searchStore)}`;
  return `
    <div class="store-map-popup">
      <strong>${shop.name}</strong>
      <div class="mt-1 small">${shop.address}</div>
      <div class="mt-2">
        <a href="${shop.website}" target="_blank" rel="noreferrer">Website</a>
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
  const onMarkerClickRef = useRef(onMarkerClick);

  useEffect(() => {
    onMarkerClickRef.current = onMarkerClick;
  }, [onMarkerClick]);

  useEffect(() => {
    if (!mapContainerRef.current || mapRef.current) {
      return;
    }

    const map = L.map(mapContainerRef.current, {
      scrollWheelZoom: true,
    }).setView(SINGAPORE_CENTER, DEFAULT_ZOOM);

    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      attribution:
        '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
      maxZoom: 19,
    }).addTo(map);

    const markers = {};
    for (const shop of stores) {
      const marker = L.marker([shop.lat, shop.lng], { icon: defaultIcon })
        .addTo(map)
        .bindPopup(buildPopupContent(shop));

      marker.on("click", () => {
        onMarkerClickRef.current?.(shop.id);
      });

      markers[shop.id] = marker;
    }

    if (stores.length > 0) {
      const bounds = L.latLngBounds(stores.map((shop) => [shop.lat, shop.lng]));
      map.fitBounds(bounds.pad(0.12));
    }

    mapRef.current = map;
    markersRef.current = markers;

    return () => {
      map.remove();
      mapRef.current = null;
      markersRef.current = {};
    };
  }, [stores]);

  useEffect(() => {
    if (!activeStoreId || !mapRef.current) {
      return;
    }

    const marker = markersRef.current[activeStoreId];
    const shop = stores.find((store) => store.id === activeStoreId);
    if (!marker || !shop) {
      return;
    }

    mapRef.current.setView([shop.lat, shop.lng], 14, { animate: true });
    marker.openPopup();
  }, [activeStoreId, stores]);

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
