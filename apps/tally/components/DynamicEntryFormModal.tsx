// vi: set ts=2 sw=2
import { useState, useEffect, useMemo, useRef } from 'react';
import { StyleSheet, ScrollView } from 'react-native';
import { Dialog, Portal, Button } from 'react-native-paper';
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
  const [selectedSchema, setSelectedSchema] = useState<Schema | null>(schema);
  const [formData, setFormData] = useState<Record<string, any>>({});
  const [errors, setErrors] = useState<Record<string, string>>({});
  const prevVisibleRef = useRef(false);

  const isLocked = lockSchema || Boolean(schema && (!schemas || schemas.length <= 1));

  // Reset form and sync schema strictly when visibility opens
  useEffect(() => {
    if (!prevVisibleRef.current && visible) {
      const active = schema || (schemas && schemas.length > 0 ? schemas[0] : null);
      setSelectedSchema(active);
      setFormData(initialData || {});
      setErrors({});
    }
    prevVisibleRef.current = visible;
  }, [visible]);

  // If a locked schema prop changes while open, sync selectedSchema
  useEffect(() => {
    if (schema) {
      setSelectedSchema(schema);
    }
  }, [schema?.tallyId]);

  // Derive parsed JSON Schema synchronously from the active raw schema string
  const currentRawSchema = selectedSchema?.jsonSchema || schema?.jsonSchema || jsonSchemaRaw;
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

  const handleValidationAndSubmit = () => {
    if (!parsedSchema) {
      if (!selectedSchema && schemas.length > 0) {
        setErrors({ _schema: 'Please select a tally schema first' });
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

    onSubmit(data, selectedSchema || undefined);
  };

  return (
    <Portal>
      <Dialog visible={visible} onDismiss={onDismiss} style={styles.dialog}>
        <Dialog.Title>{title}</Dialog.Title>
        <Dialog.ScrollArea style={styles.scrollArea}>
          <ScrollView contentContainerStyle={styles.scrollContent}>
            {/* Schema Selector or Locked Schema Display */}
            <SchemaPicker
              schemas={schemas}
              selectedSchema={selectedSchema}
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
            disabled={isSubmitting || (!selectedSchema && !parsedSchema)}
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
  },
  scrollContent: {
    paddingHorizontal: 24,
    paddingVertical: 12,
  },
});
