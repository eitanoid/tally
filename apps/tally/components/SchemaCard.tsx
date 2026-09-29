// vi: set ts=2 sw=2
import { StyleSheet, View } from 'react-native';
import { Surface, Text, TouchableRipple } from 'react-native-paper';
import { Schema } from '../generated/tally/v1/service_pb';

export interface TallyViewModel {
  schema: Schema;
  count: number;
  lastRecordedAt: string;
}

interface SchemaCardProps {
  tally: TallyViewModel;
  onPress: () => void;
  onQuickIncrement?: () => void;
}

export function SchemaCard({ tally, onPress }: SchemaCardProps) {
  return (
    <Surface style={styles.horizontalCard} elevation={1}>
      <TouchableRipple
        style={styles.cardRipple}
        onPress={onPress}
        rippleColor="rgba(0, 0, 0, .1)"
      >
        <View style={styles.cardRow}>
          {/* Left Column: Name & Description */}
          <View style={styles.leftCol}>
            <Text variant="titleMedium" style={styles.tallyName} numberOfLines={1}>
              {tally.schema.name}
            </Text>
            <Text variant="bodySmall" style={styles.tallyDescription} numberOfLines={1}>
              {tally.schema.description || 'No description'}
            </Text>
          </View>

          {/* Right Column: Count & Timestamp */}
          <View style={styles.rightCol}>
            <Text variant="headlineMedium" style={styles.tallyCount}>
              {tally.count}
            </Text>
            <Text variant="labelSmall" style={styles.lastRecorded}>
              {tally.lastRecordedAt}
            </Text>
          </View>
        </View>
      </TouchableRipple>
    </Surface>
  );
}

const styles = StyleSheet.create({
  horizontalCard: {
    marginBottom: 8,
    borderRadius: 8,
    overflow: 'hidden',
  },
  cardRipple: {
    paddingVertical: 12,
    paddingHorizontal: 16,
  },
  cardRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  leftCol: {
    flex: 1,
    paddingRight: 12,
  },
  rightCol: {
    alignItems: 'flex-end',
  },
  tallyName: {
    fontWeight: '700',
    fontSize: 16,
    lineHeight: 20,
  },
  tallyDescription: {
    opacity: 0.6,
    fontSize: 12,
    marginTop: 2,
  },
  tallyCount: {
    fontWeight: 'bold',
    fontSize: 22,
    lineHeight: 26,
  },
  lastRecorded: {
    opacity: 0.4,
    fontSize: 10,
    marginTop: 1,
  },
});
