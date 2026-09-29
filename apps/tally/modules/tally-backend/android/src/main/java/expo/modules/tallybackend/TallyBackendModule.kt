package expo.modules.tallybackend

import android.util.Log
import bridge.Bridge
import expo.modules.kotlin.modules.Module
import expo.modules.kotlin.modules.ModuleDefinition

class TallyBackendModule : Module() {
  override fun definition() = ModuleDefinition {
    Name("TallyBackend")

    OnCreate {
      val context = appContext.reactContext ?: return@OnCreate
      val dbPath = "${context.filesDir.absolutePath}/tally.db"

      try {
        Bridge.new_(dbPath)
        Log.d("TallyBackend", "Go Bridge initialized at $dbPath")
      } catch (e: Exception) {
        Log.e("TallyBackend", "Failed to initialize Go Bridge: ${e.message}", e)
      }
    }

    OnDestroy {
      try {
        Bridge.close()
        Log.d("TallyBackend", "Go Bridge stopped cleanly.")
      } catch (e: Exception) {
        Log.e("TallyBackend", "Error closing Go Bridge: ${e.message}", e)
      }
    }
    Function("hello") {
        "Hello from Android!"
    }


    Function("pingGo") { name: String ->
      Bridge.ping(name)
    }

    AsyncFunction("listEntries") { 
        reqBytes: ByteArray ->
            val respBytes = Bridge.listEntries(reqBytes)
            respBytes
    }

    AsyncFunction("createSchema") { 
        reqBytes: ByteArray ->
            val respBytes = Bridge.createSchema(reqBytes)
            respBytes
    }
    AsyncFunction("getLatestSchema") {
        reqBytes: ByteArray ->
            val respBytes = Bridge.getLatestSchema(reqBytes)
            respBytes

    }
    AsyncFunction("listSchemas"){
        reqBytes: ByteArray ->
            val respBytes = Bridge.listSchemas(reqBytes)
            respBytes
    }
    AsyncFunction("recordEntry") {
        reqBytes: ByteArray ->
            val respBytes = Bridge.recordEntry(reqBytes)
            respBytes
    }
  }
}
