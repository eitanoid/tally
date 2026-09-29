// vi: set ts=2 sw=2
import { ParsedJsonSchema, FieldInputType, Property } from './types';

export function getFieldInputType(prop: Property): FieldInputType {
  if (prop.type === 'boolean') return 'boolean';
  if (prop.type === 'integer') return 'integer';
  if (prop.type === 'number') return 'number';
  if (prop.format === 'date-time') return 'date-time';
  if (prop.format === 'date') return 'date';
  if (prop.format === 'time') return 'time';
  if (prop.format === 'go-duration' || prop.format === 'duration') return 'duration';
  return 'string';
}

export interface ValidationOutput {
  isValid: boolean;
  errors: Record<string, string>;
  data: Record<string, any>;
}

export function validateAndFormatEntryData(
  formData: Record<string, any>,
  parsedSchema: ParsedJsonSchema
): ValidationOutput {
  const errors: Record<string, string> = {};
  const data: Record<string, any> = {};

  const properties = parsedSchema.properties || {};
  const requiredList = parsedSchema.required || [];

  for (const [key, prop] of Object.entries(properties)) {
    const fieldType = getFieldInputType(prop);
    const isRequired = requiredList.includes(key);
    const rawVal = formData[key];

    // Boolean field handling
    if (fieldType === 'boolean') {
      data[key] = Boolean(rawVal);
      continue;
    }

    // Required check
    const isEmpty = rawVal === undefined || rawVal === null || String(rawVal).trim() === '';
    if (isEmpty) {
      if (isRequired) {
        errors[key] = 'This field is required';
      }
      continue;
    }

    const strVal = String(rawVal).trim();

    switch (fieldType) {
      case 'integer': {
        if (!/^-?\d+$/.test(strVal)) {
          errors[key] = 'Must be a valid integer';
        } else {
          data[key] = parseInt(strVal, 10);
        }
        break;
      }

      case 'number': {
        const num = Number(strVal);
        if (isNaN(num)) {
          errors[key] = 'Must be a valid number';
        } else {
          data[key] = num;
        }
        break;
      }

      case 'date-time': {
        const isoRegex = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/;
        if (!isoRegex.test(strVal) || isNaN(Date.parse(strVal))) {
          errors[key] = 'Must be ISO format (e.g. 2026-09-29T12:00:00Z)';
        } else {
          data[key] = strVal;
        }
        break;
      }

      case 'date': {
        const dateRegex = /^\d{4}-(0[1-9]|1[0-2])-(0[1-9]|[12]\d|3[01])$/;
        if (!dateRegex.test(strVal)) {
          errors[key] = 'Must be YYYY-MM-DD format (e.g. 2026-09-29)';
        } else {
          const [y, m, d] = strVal.split('-').map(Number);
          const dateObj = new Date(y, m - 1, d);
          if (dateObj.getMonth() !== m - 1 || dateObj.getDate() !== d) {
            errors[key] = 'Invalid calendar date';
          } else {
            data[key] = strVal;
          }
        }
        break;
      }

      case 'time': {
        const timeRegex = /^([01]\d|2[0-3]):[0-5]\d:[0-5]\d(?:Z|[+-][01]\d:[0-5]\d)?$/;
        if (!timeRegex.test(strVal)) {
          errors[key] = 'Must be HH:mm:ss format (e.g. 14:30:00)';
        } else {
          // Normalize: append Z if no timezone offset present
          const finalTime = /(?:Z|[+-]\d{2}:\d{2})$/.test(strVal) ? strVal : `${strVal}Z`;
          data[key] = finalTime;
        }
        break;
      }

      case 'duration': {
        const durationRegex = /^(?:(?:\d+(?:\.\d+)?\s*(?:h|m|s|ms|us|µs|ns)\s*)+|\d+)$/i;
        if (!durationRegex.test(strVal)) {
          errors[key] = 'Must be a duration (e.g. 1h 30m, 45s)';
        } else {
          if (/^\d+$/.test(strVal)) {
            data[key] = `${strVal}s`;
          } else {
            data[key] = strVal.replace(/\s+/g, '');
          }
        }
        break;
      }

      case 'string':
      default: {
        data[key] = strVal;
        break;
      }
    }
  }

  return {
    isValid: Object.keys(errors).length === 0,
    errors,
    data,
  };
}
