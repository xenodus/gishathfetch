/**
 * Compute a LatLngBoundsLiteral that fits all LGS marker positions.
 */
export function computeLgsMapBounds(stores) {
  return stores.reduce(
    (bounds, store) => ({
      west: Math.min(bounds.west, store.lng),
      east: Math.max(bounds.east, store.lng),
      north: Math.max(bounds.north, store.lat),
      south: Math.min(bounds.south, store.lat),
    }),
    { west: 180, east: -180, north: -90, south: 90 },
  );
}
