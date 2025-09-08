import { registerSW } from "virtual:pwa-register";

registerSW({
  immediate: true,
  onRegisteredSW(_swScriptUrl) {
    // console.log("SW registered: ", _swScriptUrl);
  },
  onOfflineReady() {
    // console.log("PWA application ready to work offline");
  },
});
