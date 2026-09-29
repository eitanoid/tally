// vi: set ts=2 sw=2
import { StyleSheet, View } from 'react-native';
import { Text, IconButton, SegmentedButtons, Surface, useTheme } from 'react-native-paper';
import { Schema } from '../generated/tally/v1/service_pb';

interface SchemaDetailHeaderProps {
  schema: Schema;
  totalEntries: number;
  viewMode: 'cards' | 'table';
  onViewModeChange: (mode: 'cards' | 'table') => void;
  onBack: () => void;
}

export function SchemaDetailHeader({
  schema,
  totalEntries,
  viewMode,
  onViewModeChange,
  onBack,
}: SchemaDetailHeaderProps) {
  const theme = useTheme();

  return (
    <View style={styles.container}>
      {/* Top App Bar with back navigation and view switcher */}
      <View style={[styles.appBar, { borderBottomColor: theme.colors.outlineVariant }]}>
        <IconButton icon="arrow-left" onPress={onBack} size={24} style={styles.backButton} />
        <Text variant="titleMedium" numberOfLines={1} style={styles.appBarTitle}>
          {schema.name}
        </Text>
        <View style={styles.toggleContainer}>
          <SegmentedButtons
            value={viewMode}
            onValueChange={(val) => onViewModeChange(val as 'cards' | 'table')}
            density="high"
            style={styles.segmentedButtons}
            buttons={[
              {
                value: 'cards',
                icon: 'view-grid-outline',
                style: styles.segmentBtn,
              },
              {
                value: 'table',
                icon: 'table',
                style: styles.segmentBtn,
              },
            ]}
          />
        </View>
      </View>

      {/* Header Metadata section: Name, Description, Total Entry Count */}
      <Surface style={[styles.metadataCard, { backgroundColor: theme.colors.elevation.level1 }]} elevation={1}>
        <View style={styles.metadataContent}>
          <View style={styles.metadataMain}>
            <Text variant="titleMedium" style={styles.metadataName} numberOfLines={1}>
              {schema.name}
            </Text>
            <Text variant="bodySmall" style={styles.metadataDescription} numberOfLines={2}>
              {schema.description || 'No description provided'}
            </Text>
          </View>
          <View style={[styles.metadataCountBadge, { backgroundColor: theme.colors.surfaceVariant }]}>
            <Text variant="headlineSmall" style={styles.metadataCountNumber}>
              {totalEntries}
            </Text>
            <Text variant="labelSmall" style={styles.metadataCountLabel}>
              {totalEntries === 1 ? 'entry' : 'entries'}
            </Text>
          </View>
        </View>
      </Surface>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    width: '100%',
  },
  appBar: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingLeft: 4,
    paddingRight: 12,
    minHeight: 56,
    borderBottomWidth: StyleSheet.hairlineWidth,
  },
  backButton: {
    margin: 0,
    marginRight: 4,
  },
  appBarTitle: {
    flex: 1,
    fontWeight: '700',
    marginRight: 12,
  },
  toggleContainer: {
    width: 108,
    flexShrink: 0,
  },
  segmentedButtons: {
    width: 108,
  },
  segmentBtn: {
    minWidth: 46,
    paddingHorizontal: 0,
  },
  metadataCard: {
    marginHorizontal: 12,
    marginTop: 8,
    marginBottom: 4,
    paddingVertical: 10,
    paddingHorizontal: 14,
    borderRadius: 8,
  },
  metadataContent: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  metadataMain: {
    flex: 1,
    paddingRight: 12,
  },
  metadataName: {
    fontWeight: '700',
    fontSize: 16,
    lineHeight: 20,
  },
  metadataDescription: {
    opacity: 0.6,
    marginTop: 2,
  },
  metadataCountBadge: {
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 8,
    minWidth: 56,
  },
  metadataCountNumber: {
    fontWeight: '800',
    lineHeight: 24,
  },
  metadataCountLabel: {
    opacity: 0.7,
    fontSize: 11,
    marginTop: 1,
  },
});
