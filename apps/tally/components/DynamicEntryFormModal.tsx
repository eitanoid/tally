// vi: set ts=2 sw=2
import { useState, useEffect, useMemo } from 'react';
import { StyleSheet, ScrollView } from 'react-native';
import { Dialog, Portal, Button, Text, useTheme } from 'react-native-paper';
import { Schema } from '../generated/tally/v1/service_pb';
import {
  DynamicEntryFormModalProps,
  ParsedJsonSchema,
  Property,
} from './entry-form/types';
import { validateAndFormatEntryData } from './entry-form/validation';
import { FieldInput } from './entry-form/FieldInput';
import { SchemaPicker } from './entry-form/SchemaPicker';

export type {
  DynamicEntryFormModalProps,
  ParsedJsonSchema,
  Property,
};

export function DynamicEntryFormModal({
  visible,
  title = 'Record Entry',
  schema = null,
  schemas = [],
  lockSchema = false,
  jsonSchemaRaw,
  initialData,
  onDismiss,
  onSubmit,
  isSubmitting,
}: DynamicEntryFormModalProps) {
  const theme = useTheme();
  const [selectedSchema, setSelectedSchema] = useState<Schema | null>(
    schema || (schemas.length > 0 ? schemas[0] : null)
  );
  const [formData, setFormData] = useState<Record<string, any>>(initialData || {});
  const [errors, setErrors] = useState<Record<string, string>>({});

  const isLocked = lockSchema || Boolean(schema && (!schemas || schemas.length <= 1));

  // Sync schema and form data whenever visible opens or schema/initialData changes
  useEffect(() => {
    if (visible) {
      const active = schema || selectedSchema || (schemas.length > 0 ? schemas[0] : null);
      setSelectedSchema(active);
      setFormData(initialData ? { ...initialData } : {});
      setErrors({});
    }
  }, [visible, schema?.tallyId, initialData]);

  // If schemas list loads asynchronously while modal is open without a selection, select the first schema
  useEffect(() => {
    if (!selectedSchema && !schema && schemas.length > 0) {
      setSelectedSchema(schemas[0]);
    }
  }, [schemas, selectedSchema, schema]);

  // Derive active schema string and parsed JSON Schema
  const currentRawSchema =
    selectedSchema?.jsonSchema || schema?.jsonSchema || jsonSchemaRaw;

  const parsedSchema = useMemo<ParsedJsonSchema | null>(() => {
    if (!currentRawSchema) return null;
    try {
      return JSON.parse(currentRawSchema);
    } catch (err) {
      console.error('Failed to parse JSON Schema:', err);
      return null;
    }
  }, [currentRawSchema]);

  const updateField = (key: string, value: any) => {
    setFormData((prev) => ({ ...prev, [key]: value }));
    if (errors[key]) {
      setErrors((prev) => ({ ...prev, [key]: '' }));
    }
  };

  const handleSelectSchema = (newSchema: Schema) => {
    setSelectedSchema(newSchema);
    setFormData({});
    setErrors({});
  };

  const handleValidationAndSubmit = async () => {
    const activeSchema = schema || selectedSchema;

    if (!parsedSchema) {
      if (!activeSchema && schemas.length > 0) {
        setErrors({ _schema: 'Please select a tally first' });
      }
      return;
    }

    const { isValid, errors: validationErrors, data } = validateAndFormatEntryData(
      formData,
      parsedSchema
    );

    if (!isValid) {
      setErrors(validationErrors);
      return;
    }

    try {
      await onSubmit(data, activeSchema || undefined);
    } catch (err) {
      console.error('Submit entry error in modal:', err);
    }
  };

  const activeSchema = schema || selectedSchema;
  const isSaveDisabled = isSubmitting || (!activeSchema && !parsedSchema);

  return (
    <Portal>
      <Dialog visible={visible} onDismiss={onDismiss} style={styles.dialog}>
        <Dialog.Title>{title}</Dialog.Title>
        <Dialog.ScrollArea style={styles.scrollArea}>
          <ScrollView
            contentContainerStyle={styles.scrollContent}
            keyboardShouldPersistTaps="handled"
          >
            {/* Schema Selector or Locked Schema Display */}
            <SchemaPicker
              schemas={schemas}
              selectedSchema={activeSchema}
              isLocked={isLocked}
              onSelectSchema={handleSelectSchema}
              error={errors._schema}
            />

            {/* Dynamic Form Inputs */}
            {parsedSchema &&
              Object.entries(parsedSchema.properties || {}).map(([key, prop]) => {
                const isRequired = parsedSchema.required?.includes(key);
                return (
                  <FieldInput
                    key={key}
                    fieldKey={key}
                    property={prop}
                    value={formData[key]}
                    error={errors[key]}
                    isRequired={isRequired}
                    onChange={updateField}
                  />
                );
              })}
            {parsedSchema && Object.keys(parsedSchema.properties || {}).length === 0 ? (
              <Text
                variant="bodyMedium"
                style={[styles.emptySchemaMessage, { color: theme.colors.onSurfaceVariant }]}
              >
                This schema has no fields. Save to record an empty entry.
              </Text>
            ) : null}
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
            disabled={isSaveDisabled}
          >
            Save Entry
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
    flexShrink: 1,
  },
  scrollContent: {
    paddingHorizontal: 24,
    paddingVertical: 12,
  },
  emptySchemaMessage: {
    paddingVertical: 16,
    textAlign: 'center',
  },
});
