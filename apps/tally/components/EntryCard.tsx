// vi: set ts=2 sw=2
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';
import { Surface, Text, IconButton, Menu, useTheme } from 'react-native-paper';
import { Entry, FieldFormat } from '../generated/tally/v1/service_pb';
import { formatCreatedAt } from '../src/utils/date';
import { formatFieldValue } from '../src/fields/fieldFormat';

interface EntryCardProps {
  entry: Entry;
  fieldFormats?: Record<string, FieldFormat>;
  onEdit?: (entry: Entry) => void;
  onDelete?: (entry: Entry) => void;
}

export function EntryCard({ entry, fieldFormats = {}, onEdit, onDelete }: EntryCardProps) {
  const theme = useTheme();
  const [menuVisible, setMenuVisible] = useState(false);

  let dataObj: Record<string, any> = {};
  try {
    dataObj = JSON.parse(entry.data || '{}');
  } catch { }

  const timeInfo = formatCreatedAt(entry.createdAt, entry.data);

  return (
    <Surface
      style={[styles.entryCard, { backgroundColor: theme.colors.elevation.level1 }]}
      elevation={1}
    >
      <View style={styles.entryHeader}>
        <View style={styles.timestampContainer}>
          <Text variant="labelSmall" style={[styles.timestampRelative, { color: theme.colors.primary }]}>
            {timeInfo.relative}
          </Text>
          <Text variant="labelSmall" style={styles.timestampDot}>
            •
          </Text>
          <Text variant="labelSmall" style={styles.timestampFormatted}>
            {timeInfo.formatted}
          </Text>
        </View>

        {(onEdit || onDelete) && (
          <Menu
            visible={menuVisible}
            onDismiss={() => setMenuVisible(false)}
            anchor={
              <IconButton
                icon="dots-vertical"
                size={18}
                onPress={() => setMenuVisible(true)}
              />
            }
          >
            {onEdit && (
              <Menu.Item
                onPress={() => {
                  setMenuVisible(false);
                  onEdit(entry);
                }}
                title="Edit"
                leadingIcon="pencil"
              />
            )}
            {onDelete && (
              <Menu.Item
                onPress={() => {
                  setMenuVisible(false);
                  onDelete(entry);
                }}
                title="Delete"
                leadingIcon="delete"
              />
            )}
          </Menu>
        )}
      </View>

      {Object.entries(dataObj).map(([key, val]) => (
        <View key={key} style={styles.dataRow}>
          <Text variant="bodyMedium" style={[styles.dataKey, { color: theme.colors.onSurface }]}>
            {key}:
          </Text>
          <Text variant="bodyMedium" style={[styles.dataVal, { color: theme.colors.onSurfaceVariant }]}>
            {formatFieldValue(val, fieldFormats[key] ?? FieldFormat.UNSPECIFIED)}
          </Text>
        </View>
      ))}
    </Surface>
  );
}

const styles = StyleSheet.create({
  entryCard: {
    padding: 12,
    borderRadius: 8,
    marginBottom: 8,
  },
  entryHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: -8,
  },
  timestampContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    flexWrap: 'wrap',
    flex: 1,
    paddingRight: 8,
  },
  timestampRelative: {
    fontWeight: '700',
  },
  timestampDot: {
    marginHorizontal: 4,
    opacity: 0.4,
  },
  timestampFormatted: {
    opacity: 0.5,
  },
  dataRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginVertical: 2,
  },
  dataKey: {
    fontWeight: '600',
    marginRight: 6,
  },
  dataVal: {
    opacity: 0.8,
  },
});
