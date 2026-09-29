// vi: set ts=2 sw=2
import { useState } from 'react';
import { StyleSheet, View, ScrollView } from 'react-native';
import {
  Dialog,
  Portal,
  TextInput,
  Button,
  Text,
  SegmentedButtons,
  IconButton,
  Switch,
  Divider,
} from 'react-native-paper';
import { FieldFormat } from '../generated/tally/v1/service_pb';

export interface DynamicField {
  id: string;
  name: string;
  description: string;
  format: FieldFormat;
  required: boolean;
}

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
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
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
    if (fields.length === 1) return; // Keep at least one field
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

            {fields.map((field, index) => (
              <View key={field.id} style={styles.fieldBlock}>
                <View style={styles.fieldHeaderRow}>
                  <Text variant="labelLarge" style={styles.fieldNumber}>
                    Field #{index + 1}
                  </Text>
                  {fields.length > 1 && (
                    <IconButton
                      icon="delete-outline"
                      size={18}
                      iconColor="red"
                      onPress={() => removeField(field.id)}
                    />
                  )}
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

                <Text variant="bodySmall" style={{ marginBottom: 4 }}>
                  Data Type:
                </Text>
                <SegmentedButtons
                  value={field.format.toString()}
                  onValueChange={(val) =>
                    updateField(field.id, { format: Number(val) as FieldFormat })
                  }
                  density="high"
                  style={{ marginBottom: 8 }}
                  buttons={[
                    { value: FieldFormat.INTEGER.toString(), label: 'Int' },
                    { value: FieldFormat.STRING.toString(), label: 'String' },
                    { value: FieldFormat.BOOLEAN.toString(), label: 'Bool' },
                  ]}
                />

                <View style={styles.switchRow}>
                  <Text variant="bodyMedium">Required Field</Text>
                  <Switch
                    value={field.required}
                    onValueChange={(val) => updateField(field.id, { required: val })}
                  />
                </View>
                <Divider style={{ marginTop: 12 }} />
              </View>
            ))}
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
  fieldHeaderRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  fieldNumber: {
    opacity: 0.7,
    fontWeight: '600',
  },
  switchRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    marginTop: 4,
  },
});
