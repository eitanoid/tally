// vi: set ts=2 sw=2
import { FieldFormat } from '../../generated/tally/v1/service_pb';

export { FieldFormat };

export interface FieldFormatInfo {
  value: FieldFormat;
  label: string;
  icon: string;
  description: string;
  placeholder?: string;
  example?: string;
}

export const FIELD_FORMAT_OPTIONS: FieldFormatInfo[] = [
  {
    value: FieldFormat.STRING,
    label: 'String',
    icon: 'format-text',
    description: 'Text input',
    placeholder: 'Enter text',
    example: 'Sample text',
  },
  {
    value: FieldFormat.INTEGER,
    label: 'Integer',
    icon: 'numeric',
    description: 'Whole number',
    placeholder: 'e.g. 42',
    example: '42',
  },
  {
    value: FieldFormat.NUMBER,
    label: 'Number',
    icon: 'decimal',
    description: 'Floating-point number',
    placeholder: 'e.g. 3.14',
    example: '3.14',
  },
  {
    value: FieldFormat.BOOLEAN,
    label: 'Boolean',
    icon: 'toggle-switch-outline',
    description: 'True / False toggle',
    example: 'true',
  },
  {
    value: FieldFormat.DATE_TIME,
    label: 'Date & Time',
    icon: 'calendar-clock',
    description: 'ISO Timestamp (YYYY-MM-DDTHH:mm:ssZ)',
    placeholder: 'YYYY-MM-DDTHH:mm:ssZ',
    example: '2026-09-29T12:00:00Z',
  },
  {
    value: FieldFormat.DATE,
    label: 'Date',
    icon: 'calendar',
    description: 'Date (YYYY-MM-DD)',
    placeholder: 'YYYY-MM-DD',
    example: '2026-09-29',
  },
  {
    value: FieldFormat.TIME,
    label: 'Time',
    icon: 'clock-outline',
    description: 'Time (HH:mm:ss)',
    placeholder: 'HH:mm:ss',
    example: '14:30:00',
  },
  {
    value: FieldFormat.DURATION,
    label: 'Duration',
    icon: 'timer-outline',
    description: 'Duration (e.g. 1h 30m or 45s)',
    placeholder: 'e.g. 1h 30m or 45s',
    example: '1h 30m',
  },
  {
    value: FieldFormat.UNSPECIFIED,
    label: 'Unspecified',
    icon: 'help-circle-outline',
    description: 'Fallback text input',
    placeholder: 'Enter value',
  },
];

export function getFieldFormatInfo(format: FieldFormat): FieldFormatInfo {
  return (
    FIELD_FORMAT_OPTIONS.find((opt) => opt.value === format) ||
    FIELD_FORMAT_OPTIONS[FIELD_FORMAT_OPTIONS.length - 1]
  );
}

export function inferFieldFormatFromProperty(prop: {
  type?: string;
  format?: string;
}): FieldFormat {
  if (prop.type === 'boolean') return FieldFormat.BOOLEAN;
  if (prop.type === 'integer') return FieldFormat.INTEGER;
  if (prop.type === 'number') return FieldFormat.NUMBER;
  if (prop.format === 'date-time') return FieldFormat.DATE_TIME;
  if (prop.format === 'date') return FieldFormat.DATE;
  if (prop.format === 'time') return FieldFormat.TIME;
  if (prop.format === 'go-duration' || prop.format === 'duration') return FieldFormat.DURATION;
  if (prop.type === 'string') return FieldFormat.STRING;
  return FieldFormat.UNSPECIFIED;
}
