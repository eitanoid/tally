// vi: set ts=2 sw=2
import { useState } from 'react';
import { StyleSheet, View } from 'react-native';
import { Button, Menu, Text, HelperText, useTheme } from 'react-native-paper';
import { Schema } from '../../generated/tally/v1/service_pb';

interface SchemaPickerProps {
  schemas: Schema[];
  selectedSchema: Schema | null;
  isLocked: boolean;
  onSelectSchema: (schema: Schema) => void;
  error?: string;
}

export function SchemaPicker({
  schemas,
  selectedSchema,
  isLocked,
  onSelectSchema,
  error,
}: SchemaPickerProps) {
  const theme = useTheme();
  const [menuVisible, setMenuVisible] = useState(false);

  if (isLocked) {
    return (
      <View style={[styles.lockedHeader, { borderBottomColor: theme.colors.outlineVariant }]}>
        <Text variant="labelSmall" style={styles.headerLabel}>
          Schema:
        </Text>
        <Text variant="titleSmall" style={styles.lockedTitle}>
          {selectedSchema?.name || 'Selected Tally'}
        </Text>
      </View>
    );
  }

  if (schemas.length === 0) {
    return (
      <View style={styles.emptyNotice}>
        <Text variant="bodyMedium" style={{ opacity: 0.6 }}>
          No tally schemas available. Please create a schema first.
        </Text>
      </View>
    );
  }

  return (
    <View style={styles.pickerContainer}>
      <Text variant="labelSmall" style={styles.headerLabel}>
        Tally Schema:
      </Text>
      <Menu
        visible={menuVisible}
        onDismiss={() => setMenuVisible(false)}
        anchor={
          <Button
            mode="outlined"
            icon="file-document-outline"
            onPress={() => setMenuVisible(true)}
            contentStyle={styles.menuAnchorContent}
            style={[styles.menuAnchorButton, { borderColor: theme.colors.outline }]}
          >
            {selectedSchema?.name || 'Select a Tally Schema...'}
          </Button>
        }
      >
        {schemas.map((s) => (
          <Menu.Item
            key={s.tallyId}
            onPress={() => {
              onSelectSchema(s);
              setMenuVisible(false);
            }}
            title={s.name}
            trailingIcon={selectedSchema?.tallyId === s.tallyId ? 'check' : undefined}
          />
        ))}
      </Menu>
      {error ? <HelperText type="error">{error}</HelperText> : null}
    </View>
  );
}

const styles = StyleSheet.create({
  lockedHeader: {
    paddingBottom: 12,
    marginBottom: 8,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: 'rgba(0, 0, 0, 0.1)',
  },
  headerLabel: {
    opacity: 0.6,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    marginBottom: 4,
  },
  lockedTitle: {
    fontWeight: '700',
  },
  pickerContainer: {
    marginBottom: 12,
  },
  menuAnchorButton: {
    borderColor: 'rgba(0, 0, 0, 0.2)',
  },
  menuAnchorContent: {
    justifyContent: 'flex-start',
  },
  emptyNotice: {
    paddingVertical: 16,
    alignItems: 'center',
  },
});
