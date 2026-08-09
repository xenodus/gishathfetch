import { useCallback, useState } from "react";
import {
  ActivityIndicator,
  FlatList,
  Linking,
  Pressable,
  StyleSheet,
  TextInput,
} from "react-native";

import SearchResultCard from "@/components/SearchResultCard";
import { Text, View } from "@/components/Themed";
import { ApiNotReadyError, fetchAutocompleteSuggestions } from "@/src/api/client";
import { MAX_SEARCH_LENGTH, MIN_SEARCH_LENGTH } from "@/src/constants";
import { mockSearchResponse } from "@/src/data/mockSearchResults";
import type { SearchResponse } from "@/src/api/types";

export default function SearchScreen() {
  const [query, setQuery] = useState("");
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [results, setResults] = useState<SearchResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [statusMessage, setStatusMessage] = useState<string | null>(null);

  const onQueryChange = useCallback(async (value: string) => {
    setQuery(value);
    setStatusMessage(null);

    if (value.trim().length < 2) {
      setSuggestions([]);
      return;
    }

    const next = await fetchAutocompleteSuggestions(value);
    setSuggestions(next.slice(0, 8));
  }, []);

  const runMockSearch = useCallback(async () => {
    const trimmed = query.trim();
    if (trimmed.length < MIN_SEARCH_LENGTH) {
      setStatusMessage(`Enter at least ${MIN_SEARCH_LENGTH} characters.`);
      return;
    }

    setLoading(true);
    setSuggestions([]);
    setStatusMessage(null);

    // Simulated latency so loading states are visible during UI work.
    await new Promise((resolve) => setTimeout(resolve, 400));

    setResults(mockSearchResponse(trimmed));
    setStatusMessage("Showing mock results — live API pending mobile auth on backend.");
    setLoading(false);
  }, [query]);

  const runLiveSearch = useCallback(async () => {
    setLoading(true);
    setStatusMessage(null);
    try {
      const { ensureSession, searchCards } = await import("@/src/api/client");
      await ensureSession();
      const response = await searchCards(query.trim(), []);
      setResults(response);
      setSuggestions([]);
    } catch (err) {
      setResults(null);
      if (err instanceof ApiNotReadyError) {
        setStatusMessage(err.message);
      } else if (err instanceof Error) {
        setStatusMessage(err.message);
      } else {
        setStatusMessage("Search failed.");
      }
    } finally {
      setLoading(false);
    }
  }, [query]);

  const openStore = useCallback((url: string) => {
    Linking.openURL(url);
  }, []);

  return (
    <View style={styles.container}>
      <Text style={styles.title}>Gishath Fetch</Text>
      <Text style={styles.subtitle}>Compare MTG singles across Singapore LGS</Text>

      <TextInput
        accessibilityLabel="Card name search"
        autoCapitalize="none"
        autoCorrect={false}
        maxLength={MAX_SEARCH_LENGTH}
        onChangeText={onQueryChange}
        placeholder="Search card name…"
        style={styles.input}
        value={query}
      />

      {suggestions.length > 0 ? (
        <View style={styles.suggestions}>
          {suggestions.map((name) => (
            <Pressable
              key={name}
              onPress={() => {
                setQuery(name);
                setSuggestions([]);
              }}
              style={styles.suggestionRow}
            >
              <Text>{name}</Text>
            </Pressable>
          ))}
        </View>
      ) : null}

      <View style={styles.actions}>
        <Pressable
          accessibilityRole="button"
          onPress={runMockSearch}
          style={[styles.button, styles.primaryButton]}
        >
          <Text style={styles.primaryButtonText}>Search (mock)</Text>
        </Pressable>
        <Pressable
          accessibilityRole="button"
          onPress={runLiveSearch}
          style={[styles.button, styles.secondaryButton]}
        >
          <Text style={styles.secondaryButtonText}>Search (live API)</Text>
        </Pressable>
      </View>

      {loading ? <ActivityIndicator style={styles.loader} /> : null}

      {statusMessage ? (
        <Text style={styles.status}>{statusMessage}</Text>
      ) : null}

      {results ? (
        <FlatList
          data={results.data}
          keyExtractor={(item, index) => `${item.src}-${item.url}-${index}`}
          renderItem={({ item }) => (
            <SearchResultCard card={item} onOpenStore={openStore} />
          )}
          style={styles.results}
          ListHeaderComponent={
            results.cardKingdomPrice?.price != null ? (
              <Text style={styles.ckPrice}>
                CK reference: US${results.cardKingdomPrice.price.toFixed(2)}
              </Text>
            ) : null
          }
        />
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
    paddingTop: 8,
  },
  title: {
    fontSize: 22,
    fontWeight: "700",
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 14,
    opacity: 0.75,
    marginBottom: 16,
  },
  input: {
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: "#999",
    borderRadius: 10,
    paddingHorizontal: 12,
    paddingVertical: 10,
    fontSize: 16,
    marginBottom: 8,
  },
  suggestions: {
    borderRadius: 10,
    marginBottom: 8,
    overflow: "hidden",
  },
  suggestionRow: {
    paddingVertical: 10,
    paddingHorizontal: 12,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: "#ddd",
  },
  actions: {
    flexDirection: "row",
    gap: 8,
    marginBottom: 8,
  },
  button: {
    flex: 1,
    borderRadius: 10,
    paddingVertical: 12,
    alignItems: "center",
  },
  primaryButton: {
    backgroundColor: "#0d6efd",
  },
  secondaryButton: {
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: "#0d6efd",
  },
  primaryButtonText: {
    color: "#fff",
    fontWeight: "600",
  },
  secondaryButtonText: {
    color: "#0d6efd",
    fontWeight: "600",
  },
  loader: {
    marginVertical: 12,
  },
  status: {
    fontSize: 13,
    opacity: 0.85,
    marginBottom: 8,
  },
  results: {
    flex: 1,
  },
  ckPrice: {
    fontSize: 13,
    marginBottom: 8,
    opacity: 0.8,
  },
});
