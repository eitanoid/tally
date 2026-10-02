// vi: set ts=2 sw=2
import React, { useState } from 'react';
import { StyleSheet, View } from 'react-native';
import { TextInput, Switch, Text, HelperText, Divider, Button, Menu, Checkbox, Chip } from 'react-native-paper';
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
  const [menuVisible, setMenuVisible] = useState(false);
  const [anchorWidth, setAnchorWidth] = useState(0);
  const fieldType: FieldInputType = getFieldInputType(property);
  const label = `${property.title || fieldKey}${isRequired ? ' *' : ''}`;
  const menuContentStyle = [
    styles.enumMenuContent,
    { minWidth: anchorWidth },
  ];

  if (fieldType === 'one-of') {
    const options = (property.enum ?? []).map(String);
    return (
      <View style={styles.fieldWrapper}>
        <Text variant="bodyLarge" style={styles.enumLabel}>{label}</Text>
        {property.description ? <HelperText type="info">{property.description}</HelperText> : null}
        <Menu
          visible={menuVisible}
          onDismiss={() => setMenuVisible(false)}
          contentStyle={menuContentStyle}
          anchor={
            <Button
              mode="outlined"
              style={styles.enumSelectButton}
              onLayout={(event) => setAnchorWidth(event.nativeEvent.layout.width)}
              onPress={() => setMenuVisible(true)}
            >
              {value ? String(value) : 'Select a value'}
            </Button>
          }
        >
          <Menu.Item
            title="Clear selection"
            disabled={value === undefined || value === null || value === ''}
            onPress={() => {
              onChange(fieldKey, undefined);
              setMenuVisible(false);
            }}
          />
          {options.map((option) => (
            <Menu.Item
              key={option}
              title={option}
              trailingIcon={value === option ? 'check' : undefined}
              onPress={() => {
                onChange(fieldKey, option);
                setMenuVisible(false);
              }}
            />
          ))}
        </Menu>
        {error ? <HelperText type="error">{error}</HelperText> : null}
        <Divider style={styles.divider} />
      </View>
    );
  }

  if (fieldType === 'many-of') {
    const selectedValues: string[] = Array.isArray(value) ? value.map(String) : [];
    const options = (property.items?.enum ?? []).map(String);
    const toggleValue = (option: string) => {
      onChange(
        fieldKey,
        selectedValues.includes(option)
          ? selectedValues.filter((selected) => selected !== option)
          : [...selectedValues, option]
      );
    };

    return (
      <View style={styles.fieldWrapper}>
        <Text variant="bodyLarge" style={styles.enumLabel}>{label}</Text>
        {property.description ? <HelperText type="info">{property.description}</HelperText> : null}
        <Menu
          visible={menuVisible}
          onDismiss={() => setMenuVisible(false)}
          contentStyle={menuContentStyle}
          anchor={
            <Button
              mode="outlined"
              style={styles.enumSelectButton}
              onLayout={(event) => setAnchorWidth(event.nativeEvent.layout.width)}
              onPress={() => setMenuVisible(true)}
              icon="chevron-down"
            >
              {selectedValues.length > 0
                ? `${selectedValues.length} selected`
                : 'Select values'}
            </Button>
          }
        >
          {options.map((option) => (
            <Checkbox.Item
              key={option}
              label={option}
              status={selectedValues.includes(option) ? 'checked' : 'unchecked'}
              onPress={() => toggleValue(option)}
              style={styles.checkboxItem}
            />
          ))}
        </Menu>
        {selectedValues.length > 0 ? (
          <View style={styles.selectedChips}>
            {selectedValues.map((selected) => (
              <Chip
                key={selected}
                compact
                onClose={() => toggleValue(selected)}
                style={styles.selectedChip}
              >
                {selected}
              </Chip>
            ))}
          </View>
        ) : null}
        {error ? <HelperText type="error">{error}</HelperText> : null}
        <Divider style={styles.divider} />
      </View>
    );
  }

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
  enumLabel: {
    marginBottom: 4,
  },
  enumSelectButton: {
    width: '100%',
  },
  enumMenuContent: {
    borderRadius: 0,
  },
  checkboxItem: {
    paddingHorizontal: 0,
  },
  selectedChips: {
    flexDirection: 'row',
    flexWrap: 'wrap',
    marginTop: 8,
  },
  selectedChip: {
    marginRight: 6,
    marginBottom: 6,
  },
  description: {
    opacity: 0.6,
  },
  divider: {
    marginTop: 8,
  },
});
