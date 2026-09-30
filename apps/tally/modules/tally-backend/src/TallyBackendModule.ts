import { NativeModule, requireNativeModule } from 'expo';

declare class TallyBackendModule extends NativeModule<{}> {
    pingGo(name: string): string;
    listEntries(reqBytes: Uint8Array): Promise<Uint8Array>;
    createSchema(reqBytes: Uint8Array): Promise<Uint8Array>;
    getLatestSchema(reqBytes: Uint8Array): Promise<Uint8Array>;
    listSchemas(reqBytes: Uint8Array): Promise<Uint8Array>;
    recordEntry(reqBytes: Uint8Array): Promise<Uint8Array>;
    deleteTally(reqBytes: Uint8Array): Promise<Uint8Array>;
    deleteEntry(reqBytes: Uint8Array): Promise<Uint8Array>;
    updateEntry(reqBytes: Uint8Array): Promise<Uint8Array>;
    close(): void
}

export default requireNativeModule<TallyBackendModule>('TallyBackend');

