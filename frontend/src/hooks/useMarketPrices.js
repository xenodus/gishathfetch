import { useCallback, useState } from "react";
import { loadMarketSnapshot } from "../utils/marketPrices";

const useMarketPrices = () => {
  const [marketCard, setMarketCard] = useState(null);
  const [marketData, setMarketData] = useState(null);
  const [marketError, setMarketError] = useState(null);
  const [marketStatus, setMarketStatus] = useState("");
  const [isMarketLoading, setIsMarketLoading] = useState(false);

  const closeMarketModal = useCallback(() => {
    setMarketCard(null);
    setMarketData(null);
    setMarketError(null);
    setMarketStatus("");
    setIsMarketLoading(false);
  }, []);

  const openMarketModal = useCallback(async (card) => {
    setMarketCard(card);
    setMarketData(null);
    setMarketError(null);
    setMarketStatus("Starting market lookup…");
    setIsMarketLoading(true);

    try {
      const snapshot = await loadMarketSnapshot(card, setMarketStatus);
      setMarketData(snapshot);
      setMarketStatus("");
    } catch (error) {
      setMarketError(
        error instanceof Error ? error.message : "Market lookup failed.",
      );
      setMarketStatus("");
    } finally {
      setIsMarketLoading(false);
    }
  }, []);

  return {
    marketCard,
    marketData,
    marketError,
    marketStatus,
    isMarketLoading,
    openMarketModal,
    closeMarketModal,
  };
};

export default useMarketPrices;
