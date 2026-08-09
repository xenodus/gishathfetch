import { StyleSheet, Pressable } from "react-native";

import { Text, View } from "@/components/Themed";
import type { SearchCard } from "@/src/api/types";

type Props = {
  card: SearchCard;
  onOpenStore?: (url: string) => void;
};

export default function SearchResultCard({ card, onOpenStore }: Props) {
  const finish = card.isFoil ? "Foil" : "Non-foil";

  return (
    <View style={styles.card}>
      <View style={styles.header}>
        <Text style={styles.store}>{card.src}</Text>
        <Text style={styles.price}>S${card.price.toFixed(2)}</Text>
      </View>
      <Text style={styles.name}>{card.name}</Text>
      <Text style={styles.meta}>
        {finish}
        {card.quality ? ` · ${card.quality}` : ""}
        {card.extraInfo ? ` · ${card.extraInfo}` : ""}
      </Text>
      {onOpenStore ? (
        <Pressable
          accessibilityRole="button"
          onPress={() => onOpenStore(card.url)}
          style={styles.linkButton}
        >
          <Text style={styles.linkText}>Open store listing</Text>
        </Pressable>
      ) : null}
    </View>
  );
}

const styles = StyleSheet.create({
  card: {
    borderRadius: 12,
    padding: 14,
    marginBottom: 10,
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: "#ccc",
  },
  header: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center",
    marginBottom: 6,
  },
  store: {
    fontWeight: "600",
    fontSize: 15,
  },
  price: {
    fontWeight: "700",
    fontSize: 16,
  },
  name: {
    fontSize: 14,
    marginBottom: 4,
  },
  meta: {
    fontSize: 12,
    opacity: 0.7,
    marginBottom: 8,
  },
  linkButton: {
    alignSelf: "flex-start",
    paddingVertical: 6,
    paddingHorizontal: 10,
    borderRadius: 8,
    backgroundColor: "#0d6efd",
  },
  linkText: {
    color: "#fff",
    fontSize: 13,
    fontWeight: "600",
  },
});
