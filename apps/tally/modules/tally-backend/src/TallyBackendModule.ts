import { NativeModule, requireNativeModule } from 'expo';

declare class TallyBackendModule extends NativeModule<{}> {
    PI: number;
    hello(): string;
    pingGo(name: string): string;
}

export default requireNativeModule<TallyBackendModule>('TallyBackend');

