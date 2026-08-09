export type SearchCard = {
  name: string;
  url: string;
  img: string;
  price: number;
  inStock: boolean;
  isFoil: boolean;
  src: string;
  quality?: string;
  extraInfo?: string;
};

export type SearchStoreError = {
  store: string;
  error: string;
  statusCode?: number;
};

export type SearchStoreStats = {
  store: string;
  itemCount: number;
  durationMs: number;
};

export type CardKingdomPrice = {
  name?: string;
  price?: number;
  url?: string;
};

export type SearchResponse = {
  data: SearchCard[];
  errors?: SearchStoreError[];
  stats?: SearchStoreStats[];
  totalDurationMs?: number;
  cardKingdomPrice?: CardKingdomPrice;
};

export type SessionResponse = {
  token: string;
  expiresAt?: string;
};
