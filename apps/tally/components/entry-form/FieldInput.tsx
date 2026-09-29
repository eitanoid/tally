// vi: set ts=2 sw=2
import React from 'react';
import { StyleSheet, View } from 'react-native';
import { TextInput, Switch, Text, HelperText, Divider } from 'react-native-paper';
import { Property, FieldInputType } from './types';
import { getFieldInputType } from './validation';

interface FieldInputProps {
  fieldKey: string;
  property: Property;
  value: any;
  error?: string;
  isRequired?: boolean;
  onChange: (key: string, value: any) => void;
}

export function FieldInput({
  fieldKey,
  property,
  value,
  error,
  isRequired,
  onChange,
}: FieldInputProps) {
  const fieldType: FieldInputType = getFieldInputType(property);
  const label = `${property.title || fieldKey}${isRequired ? ' *' : ''}`;

  const setFieldToNow = () => {
    const now = new Date();
    if (fieldType === 'date-time') {
      onChange(fieldKey, now.toISOString().replace(/\.\d{3}Z$/, 'Z'));
    } else if (fieldType === 'date') {
      onChange(fieldKey, now.toISOString().split('T')[0]);
    } else if (fieldType === 'time') {
      const timePart = now.toISOString().split('T')[1].replace('Z', '').slice(0, 8);
      onChange(fieldKey, timePart);
    }
  };

  if (fieldType === 'boolean') {
    return (
      <View style={styles.fieldWrapper}>
        <View style={styles.switchRow}>
          <View style={styles.labelCol}>
            <Text variant="bodyLarge">{label}</Text>
            {property.description ? (
              <Text variant="bodySmall" style={styles.description}>
                {property.description}
              </Text>
            ) : null}
          </View>
          <Switch
            value={Boolean(value)}
            onValueChange={(val) => onChange(fieldKey, val)}
          />
        </View>
        <Divider style={styles.divider} />
      </View>
    );
  }

  let keyboardType: 'default' | 'number-pad' | 'decimal-pad' | 'numbers-and-punctuation' = 'default';
  let placeholder = '';
  let rightIcon: React.ReactNode = null;

  switch (fieldType) {
    case 'integer':
      keyboardType = 'number-pad';
      placeholder = 'e.g. 42';
      break;
    case 'number':
      keyboardType = 'decimal-pad';
      placeholder = 'e.g. 3.14';
      break;
    case 'date-time':
      placeholder = 'YYYY-MM-DDTHH:mm:ssZ';
      rightIcon = (
        <TextInput.Icon
          icon="clock-outline"
          onPress={setFieldToNow}
        />
      );
      break;
    case 'date':
      keyboardType = 'numbers-and-punctuation';
      placeholder = 'YYYY-MM-DD';
      rightIcon = (
        <TextInput.Icon
          icon="calendar"
          onPress={setFieldToNow}
        />
      );
      break;
    case 'time':
      keyboardType = 'numbers-and-punctuation';
      placeholder = 'HH:mm:ss';
      rightIcon = (
        <TextInput.Icon
          icon="clock-outline"
          onPress={setFieldToNow}
        />
      );
      break;
    case 'duration':
      placeholder = 'e.g. 1h 30m or 45s';
      break;
    case 'string':
    default:
      keyboardType = 'default';
      break;
  }

  return (
    <View style={styles.fieldWrapper}>
      <TextInput
        label={label}
        mode="outlined"
        placeholder={placeholder}
        value={value !== undefined && value !== null ? String(value) : ''}
        keyboardType={keyboardType}
        onChangeText={(text) => onChange(fieldKey, text)}
        error={Boolean(error)}
        right={rightIcon}
      />
      {property.description ? (
        <HelperText type="info">{property.description}</HelperText>
      ) : null}
      {error ? <HelperText type="error">{error}</HelperText> : null}
    </View>
  );
}

const styles = StyleSheet.create({
  fieldWrapper: {
    marginBottom: 10,
  },
  switchRow: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: 8,
  },
  labelCol: {
    flex: 1,
    paddingRight: 12,
  },
  description: {
    opacity: 0.6,
  },
  divider: {
    marginTop: 8,
  },
});
