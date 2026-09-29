// modules/tally-backend/index.ts
import { requireNativeModule } from 'expo-modules-core';

const TallyBackend = requireNativeModule('TallyBackend');

export function hello(): string {
    return TallyBackend.hello();
}
