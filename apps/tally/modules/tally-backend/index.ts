// modules/tally-backend/index.ts
import { toBinary, fromBinary, create, DescMessage, MessageShape, MessageInitShape, } from '@bufbuild/protobuf';
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
    DeleteEntryRequestSchema,
    DeleteEntryResponseSchema,
    DeleteTallyRequestSchema,
    DeleteTallyResponseSchema,
    UpdateEntryRequestSchema,
    UpdateEntryResponseSchema,
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
    deleteTally(reqBytes: Uint8Array): Promise<Uint8Array>;
    deleteEntry(reqBytes: Uint8Array): Promise<Uint8Array>;
    updateEntry(reqBytes: Uint8Array): Promise<Uint8Array>;
}

const TallyBackend = requireNativeModule<TallyBackendNativeModule>('TallyBackend');

function createFFIHandler<
    ReqSchema extends DescMessage,
    RespSchema extends DescMessage
>(
    reqSchema: ReqSchema,
    respSchema: RespSchema,
    ffiMethod: (bytes: Uint8Array) => Promise<Uint8Array>
) {
    return async (
        request: MessageInitShape<ReqSchema> = {} as MessageInitShape<ReqSchema>
    ): Promise<MessageShape<RespSchema>> => {
        const reqMsg = create(reqSchema, request);
        const reqBytes = toBinary(reqSchema, reqMsg);
        const respBytes = await ffiMethod(reqBytes);
        return fromBinary(respSchema, respBytes) as unknown as MessageShape<RespSchema>;
    };
}


export function pingGo(name: string): string {
    return TallyBackend.pingGo(name);
}

export function closeEngine(): void {
    TallyBackend.close();
}

export const createSchema = createFFIHandler(
    CreateSchemaRequestSchema,
    CreateSchemaResponseSchema,
    TallyBackend.createSchema
);
export const getLatestSchema = createFFIHandler(
    GetLatestSchemaRequestSchema,
    GetLatestSchemaResponseSchema,
    TallyBackend.getLatestSchema
);
export const listSchemas = createFFIHandler(
    ListSchemasRequestSchema,
    ListSchemasResponseSchema,
    TallyBackend.listSchemas
);

export const recordEntry = createFFIHandler(
    RecordEntryRequestSchema,
    RecordEntryResponseSchema,
    TallyBackend.recordEntry
);

export const updateEntry = createFFIHandler(
    UpdateEntryRequestSchema,
    UpdateEntryResponseSchema,
    TallyBackend.updateEntry
);

export const deleteEntry = createFFIHandler(
    DeleteEntryRequestSchema,
    DeleteEntryResponseSchema,
    TallyBackend.deleteEntry
);

export const deleteTally = createFFIHandler(
    DeleteTallyRequestSchema,
    DeleteTallyResponseSchema,
    TallyBackend.deleteTally
);

export const listEntries = createFFIHandler(
    ListEntriesRequestSchema,
    ListEntriesResponseSchema,
    TallyBackend.deleteTally
);

export default TallyBackend;


