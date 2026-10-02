import {
    createSchema as createSchemaRPC,
    listSchemas as listSchemasRPC,
} from '../../modules/tally-backend';
import "../../modules/tally-backend"
import {
    CreateSchemaRequestSchema,
    CreateSchemaResponseSchema,
    GetLatestSchemaRequestSchema,
    GetLatestSchemaResponseSchema,
    ListEntriesRequestSchema,
    ListEntriesResponseSchema,
    ListSchemasRequestSchema,
    ListSchemasResponseSchema,
    RecordEntryRequestSchema,
    RecordEntryResponseSchema,
    type CreateSchemaResponse,
    type SchemaRequestField,
    ResponseCode,
    type GetLatestSchemaResponse,
    type ListEntriesResponse,
    type ListSchemasResponse,
    type RecordEntryResponse,
    type Schema,
    FieldFormat,
} from '../../generated/tally/v1/service_pb';

export { FieldFormat };

export interface FieldDefinition {
    name: string;
    description?: string;
    type: FieldFormat;
    required?: boolean;
    enumValues?: string[];
}

export interface CreateSchemaInput {
    name: string;
    description: string;
    fields: {
        name: string;
        description?: string;
        type: FieldFormat;
        required?: boolean;
        enumValues?: string[];
    }[];
}

export async function executeCreateSchema(
    input: CreateSchemaInput
): Promise<string> {
    // Map input fields into SchemaRequestField protobuf messages
    const pbFields: SchemaRequestField[] = input.fields.map((f) => ({
        $typeName: 'tally.v1.SchemaRequestField',
        name: f.name,
        description: f.description ?? '',
        type: f.type,
        required: f.required ?? false,
        enumValues: f.enumValues ?? [],
    }));

    // 2. Call the createSchema RPC through the FFI module
    const response = await createSchemaRPC({
        name: input.name,
        description: input.description,
        fields: pbFields,
    });

    // 3. Handle RPC Response Codes
    if (response.code !== ResponseCode.OK) {
        throw new Error(
            `CreateSchema RPC Failed [Code ${response.code}]: ${response.errorMessage || 'Unknown error'
            }`
        );
    }

    return response.tallyId;
}

export async function executeListSchemas(): Promise<Schema[]> {
    const response = await listSchemasRPC({});

    if (response.code !== ResponseCode.OK) {
        throw new Error(
            `ListSchemas RPC Failed [Code ${response.code}]: ${response.errorMessage || 'Unknown error'
            }`
        );
    }
    return response.schemas;
}
