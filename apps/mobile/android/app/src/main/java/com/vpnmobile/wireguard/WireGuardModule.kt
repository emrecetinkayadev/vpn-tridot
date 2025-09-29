package com.vpnmobile.wireguard

import com.facebook.react.bridge.Arguments
import android.os.Handler
import android.os.Looper
import com.facebook.react.bridge.Promise
import com.facebook.react.bridge.ReactApplicationContext
import com.facebook.react.bridge.ReactContextBaseJavaModule
import com.facebook.react.bridge.ReactMethod
import com.facebook.react.bridge.WritableMap
import com.facebook.react.modules.core.DeviceEventManagerModule

class WireGuardModule(reactContext: ReactApplicationContext) :
    ReactContextBaseJavaModule(reactContext) {

  private var currentState: String = "disconnected"
  private val mainHandler = Handler(Looper.getMainLooper())

  override fun getName(): String = MODULE_NAME

  @ReactMethod
  fun connect(configId: String, promise: Promise) {
    val event = buildEvent(state = "connecting", message = "Stub connect for config $configId")
    currentState = "connecting"
    sendEvent(event)
    promise.resolve(event)

    mainHandler.postDelayed({
      currentState = "connected"
      val connected = buildEvent(state = currentState, message = "Stub connected")
      sendEvent(connected)
    }, 500)
  }

  @ReactMethod
  fun disconnect(promise: Promise) {
    currentState = "disconnected"
    val event = buildEvent(state = currentState, message = "Stub disconnect")
    sendEvent(event)
    promise.resolve(null)
  }

  @ReactMethod
  fun getCurrentState(promise: Promise) {
    promise.resolve(currentState)
  }

  @ReactMethod
  fun addListener(@Suppress("UNUSED_PARAMETER") eventName: String) {
    // Required for RN event emitter compatibility.
  }

  @ReactMethod
  fun removeListeners(@Suppress("UNUSED_PARAMETER") count: Int) {
    // Required for RN event emitter compatibility.
  }

  private fun buildEvent(state: String, message: String?): WritableMap {
    val map = Arguments.createMap()
    map.putString("state", state)
    map.putDouble("timestamp", System.currentTimeMillis().toDouble())
    message?.let { map.putString("message", it) }
    return map
  }

  private fun sendEvent(event: WritableMap) {
    reactApplicationContext
        .getJSModule(DeviceEventManagerModule.RCTDeviceEventEmitter::class.java)
        .emit("stateChanged", event)
  }

  companion object {
    private const val MODULE_NAME = "WireGuard"
  }
}
