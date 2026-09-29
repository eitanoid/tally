import { NativeModule, requireNativeModule } from 'expo';

declare class TallyBackendModule extends NativeModule<{}> {
  PI: number;
  hello(): string;
}

export default requireNativeModule<TallyBackendModule>('TallyBackend');
