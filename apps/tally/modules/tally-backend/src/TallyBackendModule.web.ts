import { registerWebModule, NativeModule } from 'expo';

class TallyBackendModule extends NativeModule<{}> {
  PI = Math.PI;

  hello() {
    return 'Hello world! 👋';
  }
}

export default registerWebModule(TallyBackendModule, 'TallyBackendModule');
