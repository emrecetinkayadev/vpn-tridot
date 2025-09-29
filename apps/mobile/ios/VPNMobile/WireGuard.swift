import Foundation
import React

@objc(WireGuard)
final class WireGuard: RCTEventEmitter {
  private var currentState: String = "disconnected"

  override static func requiresMainQueueSetup() -> Bool {
    false
  }

  override func supportedEvents() -> [String]! {
    ["stateChanged", "error"]
  }

  @objc(connect:resolver:rejecter:)
  func connect(_ configId: String,
               resolver: @escaping RCTPromiseResolveBlock,
               rejecter _: @escaping RCTPromiseRejectBlock) {
    let event = makeEvent(state: "connecting", message: "Stub connect for config \(configId)")
    currentState = "connecting"
    sendEvent(withName: "stateChanged", body: event)
    resolver(event)

    DispatchQueue.main.asyncAfter(deadline: .now() + 0.5) { [weak self] in
      guard let self else { return }
      self.currentState = "connected"
      let connected = self.makeEvent(state: "connected", message: "Stub connected")
      self.sendEvent(withName: "stateChanged", body: connected)
    }
  }

  @objc(disconnect:rejecter:)
  func disconnect(_ resolver: @escaping RCTPromiseResolveBlock,
                  rejecter _: @escaping RCTPromiseRejectBlock) {
    currentState = "disconnected"
    let event = makeEvent(state: currentState, message: "Stub disconnect")
    sendEvent(withName: "stateChanged", body: event)
    resolver(nil)
  }

  @objc(getCurrentState:rejecter:)
  func getCurrentState(_ resolver: @escaping RCTPromiseResolveBlock,
                       rejecter _: @escaping RCTPromiseRejectBlock) {
    resolver(currentState)
  }

  private func makeEvent(state: String, message: String?) -> [String: Any] {
    var payload: [String: Any] = [
      "state": state,
      "timestamp": Date().timeIntervalSince1970 * 1000,
    ]
    if let message = message {
      payload["message"] = message
    }
    return payload
  }
}
