import { StyleSheet } from "react-native";

import { Text, View } from "@/components/Themed";

export default function SavedScreen() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Saved cards</Text>
      <Text style={styles.body}>
        Cart persistence will use AsyncStorage, mirroring the web app&apos;s saved
        list and export/import codes. This tab is a placeholder in the initial scaffold.
      </Text>
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
    marginBottom: 12,
  },
  body: {
    fontSize: 15,
    lineHeight: 22,
    opacity: 0.85,
  },
});
