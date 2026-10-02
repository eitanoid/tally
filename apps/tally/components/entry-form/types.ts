// vi: set ts=2 sw=2
import { Schema } from '../../generated/tally/v1/service_pb';

export interface Property {
  type?: string;
  format?: string;
  title?: string;
  description?: string;
  enum?: (string | number)[];
  items?: { enum?: (string | number)[] };
  minItems?: number;
}

export interface ParsedJsonSchema {
  properties: Record<string, Property>;
  required?: string[];
  description?: string;
}

export type FieldInputType =
  | 'boolean'
  | 'integer'
  | 'number'
  | 'date-time'
  | 'date'
  | 'time'
  | 'duration'
  | 'one-of'
  | 'many-of'
  | 'string';

export interface DynamicEntryFormModalProps {
  visible: boolean;
  title?: string;
  schema?: Schema | null;
  schemas?: Schema[];
  lockSchema?: boolean;
  jsonSchemaRaw?: string;
  initialData?: Record<string, any>;
  onDismiss: () => void;
  onSubmit: (formData: Record<string, any>, selectedSchema?: Schema) => Promise<void>;
  isSubmitting: boolean;
}
