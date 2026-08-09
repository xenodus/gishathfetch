import { Linking, StyleSheet } from "react-native";

import { Text, View } from "@/components/Themed";
import { SITE_BASE_URL } from "@/src/config";

export default function AboutModal() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Gishath Fetch</Text>
      <Text style={styles.body}>
        Native app scaffold. Full search, cart, and filters are in progress.
      </Text>
      <Text
        accessibilityRole="link"
        onPress={() => Linking.openURL(SITE_BASE_URL)}
        style={styles.link}
      >
        {SITE_BASE_URL}
      </Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    padding: 20,
  },
  title: {
    fontSize: 20,
    fontWeight: "700",
    marginBottom: 12,
  },
  body: {
    fontSize: 15,
    lineHeight: 22,
    marginBottom: 16,
  },
  link: {
    color: "#0d6efd",
    fontSize: 15,
  },
});
