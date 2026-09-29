// vi: set ts=2 sw=2
import { useState, useEffect } from 'react';
import { StyleSheet, View, ScrollView } from 'react-native';
import {
  Dialog,
  Portal,
  TextInput,
  Button,
  Switch,
  Text,
  HelperText,
} from 'react-native-paper';

interface Property {
  type: string;
  title?: string;
  description?: string;
}

export interface ParsedJsonSchema {
  properties: Record<string, Property>;
  required?: string[];
}

interface DynamicEntryFormModalProps {
  visible: boolean;
  title: string;
  jsonSchemaRaw: string;
  initialData?: Record<string, any>;
  onDismiss: () => void;
  onSubmit: (formData: Record<string, any>) => Promise<void>;
  isSubmitting: boolean;
}

export function DynamicEntryFormModal({
  visible,
  title,
  jsonSchemaRaw,
  initialData = {},
  onDismiss,
  onSubmit,
  isSubmitting,
}: DynamicEntryFormModalProps) {
  const [formData, setFormData] = useState<Record<string, any>>(initialData);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [parsedSchema, setParsedSchema] = useState<ParsedJsonSchema | null>(null);

  useEffect(() => {
    if (jsonSchemaRaw) {
      try {
        const schemaObj = JSON.parse(jsonSchemaRaw);
        setParsedSchema(schemaObj);
      } catch (err) {
        console.error('Failed to parse JSON Schema:', err);
      }
    }
  }, [jsonSchemaRaw]);

  useEffect(() => {
    setFormData(initialData || {});
  }, [initialData, visible]);

  const updateField = (key: string, value: any) => {
    setFormData((prev) => ({ ...prev, [key]: value }));
    if (errors[key]) setErrors((prev) => ({ ...prev, [key]: '' }));
  };

  const handleValidationAndSubmit = () => {
    if (!parsedSchema) return;

    const newErrors: Record<string, string> = {};
    parsedSchema.required?.forEach((key) => {
      const val = formData[key];
      if (val === undefined || val === null || val === '') {
        newErrors[key] = 'This field is required';
      }
    });

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    onSubmit(formData);
  };

  return (
    <Portal>
      <Dialog visible={visible} onDismiss={onDismiss}>
        <Dialog.Title>{title}</Dialog.Title>
        <Dialog.ScrollArea>
          <ScrollView contentContainerStyle={{ paddingHorizontal: 24, paddingVertical: 8 }}>
            {parsedSchema &&
              Object.entries(parsedSchema.properties || {}).map(([key, prop]) => {
                const isRequired = parsedSchema.required?.includes(key);
                const label = `${prop.title || key}${isRequired ? ' *' : ''}`;

                if (prop.type === 'boolean') {
                  return (
                    <View key={key} style={styles.switchRow}>
                      <Text variant="bodyMedium">{label}</Text>
                      <Switch
                        value={!!formData[key]}
                        onValueChange={(val) => updateField(key, val)}
                      />
                    </View>
                  );
                }

                return (
                  <View key={key} style={{ marginBottom: 8 }}>
                    <TextInput
                      label={label}
                      mode="outlined"
                      value={formData[key] !== undefined ? String(formData[key]) : ''}
                      keyboardType={prop.type === 'integer' ? 'number-pad' : 'default'}
                      onChangeText={(text) => {
                        const val =
                          prop.type === 'integer'
                            ? text === ''
                              ? ''
                              : parseInt(text, 10)
                            : text;
                        updateField(key, val);
                      }}
                      error={!!errors[key]}
                    />
                    {prop.description && (
                      <HelperText type="info">{prop.description}</HelperText>
                    )}
                    {errors[key] && <HelperText type="error">{errors[key]}</HelperText>}
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
            onPress={handleValidationAndSubmit}
            loading={isSubmitting}
            disabled={isSubmitting}
          >
            Save Entry
          </Button>
        </Dialog.Actions>
      </Dialog>
    </Portal>
  );
}

const styles = StyleSheet.create({
  switchRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 8,
  },
});
