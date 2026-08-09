import { Linking, Pressable, StyleSheet } from "react-native";

import { Text, View } from "@/components/Themed";
import { SITE_BASE_URL } from "@/src/config";
import { SITE_TAGLINE } from "@/src/constants";

export default function FaqScreen() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>About</Text>
      <Text style={styles.body}>{SITE_TAGLINE}</Text>
      <Text style={styles.sectionTitle}>FAQ (placeholder)</Text>
      <Text style={styles.body}>
        Full FAQ content will mirror the web modals. For now, visit the website for
        pricing notes, store coverage, and privacy details.
      </Text>
      <Pressable
        accessibilityRole="link"
        onPress={() => Linking.openURL(SITE_BASE_URL)}
        style={styles.linkButton}
      >
        <Text style={styles.linkText}>Open gishathfetch.com</Text>
      </Pressable>
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
    marginBottom: 8,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: "600",
    marginTop: 16,
    marginBottom: 8,
  },
  body: {
    fontSize: 15,
    lineHeight: 22,
    opacity: 0.85,
  },
  linkButton: {
    marginTop: 20,
    backgroundColor: "#0d6efd",
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderRadius: 10,
    alignSelf: "flex-start",
  },
  linkText: {
    color: "#fff",
    fontWeight: "600",
  },
});
