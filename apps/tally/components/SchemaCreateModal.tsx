// vi: set ts=2 sw=2
import { useState } from 'react';
import { StyleSheet, View, ScrollView } from 'react-native';
import {
  Dialog,
  Portal,
  TextInput,
  Button,
  Text,
  IconButton,
  Switch,
  Divider,
  Menu,
  useTheme,
} from 'react-native-paper';
import { FieldFormat } from '../generated/tally/v1/service_pb';
import {
  FIELD_FORMAT_OPTIONS,
  getFieldFormatInfo,
  type FieldFormatInfo,
} from '../src/fields/fieldFormat';

export interface DynamicField {
  id: string;
  name: string;
  description: string;
  format: FieldFormat;
  required: boolean;
}

export type { FieldFormatInfo };

interface SchemaCreateModalProps {
  visible: boolean;
  onDismiss: () => void;
  onSubmit: (name: string, description: string, fields: DynamicField[]) => Promise<void>;
  isSubmitting: boolean;
}

export function SchemaCreateModal({
  visible,
  onDismiss,
  onSubmit,
  isSubmitting,
}: SchemaCreateModalProps) {
  const theme = useTheme();
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [openMenuFieldId, setOpenMenuFieldId] = useState<string | null>(null);
  const [fields, setFields] = useState<DynamicField[]>([
    {
      id: '1',
      name: 'count',
      description: 'Primary counter',
      format: FieldFormat.INTEGER,
      required: true,
    },
  ]);

  const addField = () => {
    setFields((prev) => [
      ...prev,
      {
        id: Date.now().toString(),
        name: '',
        description: '',
        format: FieldFormat.STRING,
        required: false,
      },
    ]);
  };

  const removeField = (id: string) => {
    setFields((prev) => prev.filter((f) => f.id !== id));
  };

  const updateField = (id: string, updates: Partial<DynamicField>) => {
    setFields((prev) => prev.map((f) => (f.id === id ? { ...f, ...updates } : f)));
  };

  const handleCreate = () => {
    if (!name.trim()) return;
    onSubmit(name.trim(), description.trim(), fields);
  };

  return (
    <Portal>
      <Dialog visible={visible} onDismiss={onDismiss} style={styles.dialog}>
        <Dialog.Title>Create Tally Schema</Dialog.Title>
        <Dialog.ScrollArea style={styles.scrollArea}>
          <ScrollView contentContainerStyle={styles.scrollContent}>
            <TextInput
              label="Schema Name *"
              value={name}
              onChangeText={setName}
              mode="outlined"
              style={styles.input}
              autoFocus
            />
            <TextInput
              label="Description (optional)"
              value={description}
              onChangeText={setDescription}
              mode="outlined"
              style={styles.input}
            />

            <View style={styles.fieldsHeader}>
              <Text variant="titleMedium" style={{ fontWeight: '700' }}>
                Fields ({fields.length})
              </Text>
              <Button mode="text" icon="plus" onPress={addField}>
                Add Field
              </Button>
            </View>

            {fields.length === 0 ? (
              <Text variant="bodyMedium" style={styles.emptyFields}>
                No fields yet. You can create an empty schema now, or add fields whenever you’re
                ready to collect structured data.
              </Text>
            ) : null}

            {fields.map((field, index) => {
              const currentOption = getFieldFormatInfo(field.format);
              return (
                <View key={field.id} style={styles.fieldBlock}>
                  <View style={styles.fieldHeaderRow}>
                    <Text variant="labelLarge" style={styles.fieldNumber}>
                      Field #{index + 1}
                    </Text>
                    <IconButton
                      icon="delete-outline"
                      size={18}
                      iconColor="red"
                      accessibilityLabel={`Remove field ${field.name || index + 1}`}
                      onPress={() => removeField(field.id)}
                    />
                  </View>

                  <TextInput
                    label="Field Name *"
                    value={field.name}
                    onChangeText={(val) => updateField(field.id, { name: val })}
                    mode="outlined"
                    dense
                    style={styles.input}
                  />
                  <TextInput
                    label="Description"
                    value={field.description}
                    onChangeText={(val) => updateField(field.id, { description: val })}
                    mode="outlined"
                    dense
                    style={styles.input}
                  />

                  <View style={styles.formatSelectorContainer}>
                    <Text variant="bodySmall" style={styles.formatLabel}>
                      Field Format:
                    </Text>
                    <Menu
                      visible={openMenuFieldId === field.id}
                      onDismiss={() => setOpenMenuFieldId(null)}
                      anchor={
                        <Button
                          mode="outlined"
                          onPress={() => setOpenMenuFieldId(field.id)}
                          icon={currentOption.icon}
                          contentStyle={styles.menuAnchorContent}
                          style={[styles.menuAnchorButton, { borderColor: theme.colors.outline }]}
                        >
                          {currentOption.label} ({currentOption.description})
                        </Button>
                      }
                    >
                      {FIELD_FORMAT_OPTIONS.map((opt) => (
                        <Menu.Item
                          key={opt.value}
                          onPress={() => {
                            updateField(field.id, { format: opt.value });
                            setOpenMenuFieldId(null);
                          }}
                          title={`${opt.label} (${opt.description})`}
                          leadingIcon={opt.icon}
                          trailingIcon={field.format === opt.value ? 'check' : undefined}
                        />
                      ))}
                    </Menu>
                  </View>

                  <View style={styles.switchRow}>
                    <Text variant="bodyMedium">Required Field</Text>
                    <Switch
                      value={field.required}
                      onValueChange={(val) => updateField(field.id, { required: val })}
                    />
                  </View>
                  <Divider style={{ marginTop: 12 }} />
                </View>
              );
            })}
          </ScrollView>
        </Dialog.ScrollArea>

        <Dialog.Actions>
          <Button onPress={onDismiss} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button
            mode="contained"
            onPress={handleCreate}
            disabled={!name.trim() || fields.some((f) => !f.name.trim()) || isSubmitting}
            loading={isSubmitting}
          >
            Create
          </Button>
        </Dialog.Actions>
      </Dialog>
    </Portal>
  );
}

const styles = StyleSheet.create({
  dialog: {
    maxHeight: '85%',
  },
  scrollArea: {
    paddingHorizontal: 0,
  },
  scrollContent: {
    paddingHorizontal: 24,
    paddingBottom: 8,
  },
  input: {
    marginBottom: 8,
  },
  fieldsHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: 8,
    marginBottom: 4,
  },
  fieldBlock: {
    marginBottom: 8,
  },
  emptyFields: {
    marginVertical: 12,
    opacity: 0.7,
  },
  fieldHeaderRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  fieldNumber: {
    opacity: 0.7,
    fontWeight: '600',
  },
  formatSelectorContainer: {
    marginBottom: 8,
  },
  formatLabel: {
    marginBottom: 4,
    opacity: 0.7,
  },
  menuAnchorButton: {
    width: '100%',
    borderColor: 'rgba(0, 0, 0, 0.2)',
  },
  menuAnchorContent: {
    justifyContent: 'flex-start',
  },
  switchRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: 4,
  },
});
