import ExpoModulesCore

public class TallyBackendModule: Module {
  public func definition() -> ModuleDefinition {
    Name("TallyBackend")

    Constant("PI") {
      Double.pi
    }

    Function("hello") {
      return "Hello world! 👋"
    }
  }
}
