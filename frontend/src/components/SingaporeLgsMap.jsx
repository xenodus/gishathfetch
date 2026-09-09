import {
  APIProvider,
  Map as GoogleMap,
  Marker,
  useMap,
} from "@vis.gl/react-google-maps";
import { useEffect } from "react";
import { GOOGLE_MAPS_API_KEY } from "../constants";
import { computeLgsMapBounds } from "../utils/lgsMapBounds";

function MapFocus({ selectedStoreId, stores }) {
  const map = useMap();

  useEffect(() => {
    if (!map || !selectedStoreId) {
      return;
    }

    const store = stores.find((entry) => entry.id === selectedStoreId);
    if (!store) {
      return;
    }

    map.panTo({ lat: store.lat, lng: store.lng });
    map.setZoom(16);
  }, [map, selectedStoreId, stores]);

  return null;
}

const SingaporeLgsMap = ({
  stores,
  selectedStoreId,
  onMarkerClick,
  isActive,
}) => {
  if (!GOOGLE_MAPS_API_KEY) {
    return (
      <div className="singapore-lgs-map singapore-lgs-map--unavailable border border-dark d-flex align-items-center justify-content-center text-muted">
        Map unavailable. Set <code>VITE_GOOGLE_MAPS_API_KEY</code> to enable the
        store map.
      </div>
    );
  }

  if (!isActive) {
    return (
      <div className="singapore-lgs-map border border-dark d-flex align-items-center justify-content-center text-muted">
        Loading map…
      </div>
    );
  }

  const defaultBounds = {
    ...computeLgsMapBounds(stores),
    padding: 48,
  };

  return (
    <div className="singapore-lgs-map border border-dark">
      <APIProvider apiKey={GOOGLE_MAPS_API_KEY}>
        <GoogleMap
          defaultBounds={defaultBounds}
          gestureHandling="cooperative"
          mapTypeControl={false}
          streetViewControl={false}
          fullscreenControl
        >
          <MapFocus selectedStoreId={selectedStoreId} stores={stores} />
          {stores.map((shop) => (
            <Marker
              key={shop.id}
              position={{ lat: shop.lat, lng: shop.lng }}
              title={shop.name}
              zIndex={selectedStoreId === shop.id ? 2 : 1}
              onClick={() => onMarkerClick?.(shop.id)}
            />
          ))}
        </GoogleMap>
      </APIProvider>
    </div>
  );
};

export default SingaporeLgsMap;
