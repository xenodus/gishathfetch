import { FlatList, Linking, Pressable, StyleSheet } from "react-native";

import { Text, View } from "@/components/Themed";
import { LGS_MAP } from "@/src/constants";

function mapsUrlForAddress(address: string): string {
  return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(address)}`;
}

export default function MapScreen() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Store map</Text>
      <Text style={styles.subtitle}>
        Tap a store for Google Maps or the shop website.
      </Text>
      <FlatList
        data={LGS_MAP}
        keyExtractor={(item) => item.id}
        renderItem={({ item }) => (
          <View style={styles.card}>
            <Text style={styles.storeName}>{item.name}</Text>
            <Text style={styles.address}>{item.address}</Text>
            <View style={styles.actions}>
              <Pressable
                accessibilityRole="button"
                onPress={() => Linking.openURL(mapsUrlForAddress(item.address))}
                style={styles.button}
              >
                <Text style={styles.buttonText}>Maps</Text>
              </Pressable>
              <Pressable
                accessibilityRole="button"
                onPress={() => Linking.openURL(item.website)}
                style={styles.button}
              >
                <Text style={styles.buttonText}>Website</Text>
              </Pressable>
            </View>
          </View>
        )}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 16,
  },
  title: {
    fontSize: 20,
    fontWeight: "700",
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 14,
    opacity: 0.75,
    marginBottom: 12,
  },
  card: {
    padding: 12,
    marginBottom: 10,
    borderRadius: 10,
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: "#ccc",
  },
  storeName: {
    fontWeight: "600",
    fontSize: 16,
    marginBottom: 4,
  },
  address: {
    fontSize: 14,
    opacity: 0.8,
    marginBottom: 8,
  },
  actions: {
    flexDirection: "row",
    gap: 8,
  },
  button: {
    backgroundColor: "#0d6efd",
    paddingVertical: 6,
    paddingHorizontal: 12,
    borderRadius: 8,
  },
  buttonText: {
    color: "#fff",
    fontWeight: "600",
    fontSize: 13,
  },
});
