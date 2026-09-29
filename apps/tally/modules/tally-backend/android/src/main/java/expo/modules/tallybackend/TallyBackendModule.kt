package expo.modules.tallybackend

import expo.modules.kotlin.modules.Module
import expo.modules.kotlin.modules.ModuleDefinition

class TallyBackendModule : Module() {
  override fun definition() = ModuleDefinition {
    Name("TallyBackend")

    Constant("PI") {
      Math.PI
    }

    Function("hello") {
      "Hello world! 👋"
    }
  }
}
