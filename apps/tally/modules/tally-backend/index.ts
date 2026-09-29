// modules/tally-backend/index.ts
import { toBinary, fromBinary, create, type DescMessage } from '@bufbuild/protobuf';
import { requireNativeModule } from 'expo-modules-core';

// 1. Import Protobuf Schemas & Types
// Adjust the import path relative to your generated protobuf directory
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
    type GetLatestSchemaResponse,
    type ListEntriesResponse,
    type ListSchemasResponse,
    type RecordEntryResponse,
} from '../../generated/tally/v1/service_pb';


// 2. Define the Native Module Interface for raw Uint8Array FFI calls
interface TallyBackendNativeModule {
    pingGo(name: string): string;
    close(): void;
    listEntries(reqBytes: Uint8Array): Promise<Uint8Array>;
    createSchema(reqBytes: Uint8Array): Promise<Uint8Array>;
    getLatestSchema(reqBytes: Uint8Array): Promise<Uint8Array>;
    listSchemas(reqBytes: Uint8Array): Promise<Uint8Array>;
    recordEntry(reqBytes: Uint8Array): Promise<Uint8Array>;
}

const TallyBackend = requireNativeModule<TallyBackendNativeModule>('TallyBackend');

// 3. Export Utility Native Methods
export function pingGo(name: string): string {
    return TallyBackend.pingGo(name);
}

export function closeEngine(): void {
    TallyBackend.close();
}

// 4. Export Typed FFI RPC Wrappers (Serialization -> Native Call -> Deserialization)
type Init<T extends DescMessage> = Parameters<typeof create<T>>[1];

export async function createSchema(
    request: Init<typeof CreateSchemaRequestSchema>
): Promise<CreateSchemaResponse> {
    const msg = create(CreateSchemaRequestSchema, request);
    const reqBytes = toBinary(CreateSchemaRequestSchema, msg);
    const respBytes = await TallyBackend.createSchema(reqBytes);
    return fromBinary(CreateSchemaResponseSchema, respBytes);
}

export async function getLatestSchema(
    request: Init<typeof GetLatestSchemaRequestSchema>
): Promise<GetLatestSchemaResponse> {
    const msg = create(GetLatestSchemaRequestSchema, request);
    const reqBytes = toBinary(GetLatestSchemaRequestSchema, msg);
    const respBytes = await TallyBackend.getLatestSchema(reqBytes);
    return fromBinary(GetLatestSchemaResponseSchema, respBytes);
}

export async function listSchemas(
    request: Init<typeof ListSchemasRequestSchema> = {}
): Promise<ListSchemasResponse> {
    const msg = create(ListSchemasRequestSchema, request);
    const reqBytes = toBinary(ListSchemasRequestSchema, msg);
    const respBytes = await TallyBackend.listSchemas(reqBytes);
    return fromBinary(ListSchemasResponseSchema, respBytes);
}

export async function recordEntry(
    request: Init<typeof RecordEntryRequestSchema>
): Promise<RecordEntryResponse> {
    const msg = create(RecordEntryRequestSchema, request);
    const reqBytes = toBinary(RecordEntryRequestSchema, msg);
    const respBytes = await TallyBackend.recordEntry(reqBytes);
    return fromBinary(RecordEntryResponseSchema, respBytes);
}

export async function listEntries(
    request: Init<typeof ListEntriesRequestSchema>
): Promise<ListEntriesResponse> {
    const msg = create(ListEntriesRequestSchema, request);
    const reqBytes = toBinary(ListEntriesRequestSchema, msg);
    const respBytes = await TallyBackend.listEntries(reqBytes);
    return fromBinary(ListEntriesResponseSchema, respBytes);
}

export default TallyBackend;


