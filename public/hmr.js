(function () {
  const DEFAULT_TIMEOUT = 1000;
  const HOT_RELOAD_URL = `ws://${location.hostname}:${location.port}/ws/hotreload`;
  let lastUuid = "";
  let timeout;

  function bumpTimeout() {
    if (timeout > 10 * DEFAULT_TIMEOUT) {
      return;
    }
    timeout += DEFAULT_TIMEOUT;
  }

  function connectHotReload() {
    const socket = new WebSocket(HOT_RELOAD_URL);

    socket.onmessage = (event) => {
      if (lastUuid === "") {
        lastUuid = event.data;
      } else if (lastUuid !== event.data) {
        console.log("[Hot Reloader] Server Changed, reloading");
        location.reload();
      }
    };

    socket.onopen = () => {
      console.log("[Hot Reloader]: Server connected");
      timeout = DEFAULT_TIMEOUT;
    };

    socket.onclose = () => {
      console.log("[Hot Reloader]: Lost Connection with server, reconnecting");
      setTimeout(function () {
        bumpTimeout();
        connectHotReload();
      }, timeout);
    };
  }

  connectHotReload();
})();
