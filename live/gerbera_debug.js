(function() {
  var host = document.createElement("div");
  host.id = "gerbera-debug-host";
  host.style.cssText = "position:fixed;bottom:0;right:0;z-index:2147483647;";
  document.body.appendChild(host);
  var shadow = host.attachShadow({mode:"closed"});

  var style = document.createElement("style");
  style.textContent = [
    ".gd-panel { font-family:monospace; font-size:12px; color:#e0e0e0;",
    "  background:#1e1e2e; border:1px solid #444; border-radius:8px 0 0 0;",
    "  width:420px; max-height:70vh; display:flex; flex-direction:column;",
    "  box-shadow:0 0 20px rgba(0,0,0,0.5); }",
    ".gd-panel.gd-collapsed { display:none; }",
    ".gd-header { display:flex; justify-content:space-between; align-items:center;",
    "  padding:8px 12px; background:#2d2d44; border-radius:8px 0 0 0; cursor:pointer; }",
    ".gd-header-title { font-weight:bold; color:#cba6f7; }",
    ".gd-tabs { display:flex; border-bottom:1px solid #444; }",
    ".gd-tab { padding:6px 12px; cursor:pointer; border:none; background:none;",
    "  color:#888; font-family:monospace; font-size:12px; }",
    ".gd-tab:hover { color:#cdd6f4; }",
    ".gd-tab.active { color:#cba6f7; border-bottom:2px solid #cba6f7; }",
    ".gd-content { overflow-y:auto; padding:8px 12px; flex:1; max-height:50vh; }",
    ".gd-entry { padding:4px 0; border-bottom:1px solid #333; word-break:break-all; }",
    ".gd-time { color:#888; }",
    ".gd-event-name { color:#89b4fa; font-weight:bold; }",
    ".gd-payload { color:#a6adc8; }",
    ".gd-badge { display:inline-block; padding:1px 6px; border-radius:4px;",
    "  font-size:10px; margin-left:4px; }",
    ".gd-badge-ok { background:#2d4a2d; color:#a6e3a1; }",
    ".gd-badge-warn { background:#4a3d2d; color:#fab387; }",
    ".gd-json { white-space:pre-wrap; word-break:break-all; color:#cdd6f4;",
    "  background:#181825; padding:8px; border-radius:4px; margin:4px 0;",
    "  max-height:40vh; overflow-y:auto; }",
    ".gd-toggle { width:36px; height:36px; border-radius:50%; border:none;",
    "  background:#cba6f7; color:#1e1e2e; font-size:16px; font-weight:bold;",
    "  cursor:pointer; box-shadow:0 2px 8px rgba(0,0,0,0.3);",
    "  display:flex; align-items:center; justify-content:center; margin:8px; }",
    ".gd-toggle:hover { background:#b48ef0; }",
    ".gd-status { display:inline-block; width:8px; height:8px; border-radius:50%;",
    "  margin-right:6px; }",
    ".gd-status-connected { background:#a6e3a1; }",
    ".gd-status-disconnected { background:#f38ba8; }",
    ".gd-session-item { padding:6px 0; border-bottom:1px solid #333; }",
    ".gd-session-label { color:#888; display:inline-block; width:100px; }",
    ".gd-session-value { color:#cdd6f4; }",
    ".gd-empty { color:#666; font-style:italic; padding:16px 0; text-align:center; }"
  ].join("\n");
  shadow.appendChild(style);

  var toggleBtn = document.createElement("button");
  toggleBtn.className = "gd-toggle";
  toggleBtn.textContent = "G";
  toggleBtn.title = "Toggle Gerbera Debug Panel (Ctrl+Shift+D)";
  shadow.appendChild(toggleBtn);

  var panel = document.createElement("div");
  panel.className = "gd-panel gd-collapsed";

  var header = document.createElement("div");
  header.className = "gd-header";
  var headerTitle = document.createElement("span");
  headerTitle.className = "gd-header-title";
  headerTitle.textContent = "Gerbera Debug";
  var headerStatus = document.createElement("span");
  headerStatus.className = "gd-status gd-status-connected";
  header.appendChild(headerTitle);
  header.appendChild(headerStatus);
  panel.appendChild(header);

  var tabBar = document.createElement("div");
  tabBar.className = "gd-tabs";
  var TABS = ["Events", "Patches", "State", "Session"];
  var tabEls = [];
  TABS.forEach(function(name, i) {
    var tab = document.createElement("button");
    tab.className = "gd-tab" + (i === 0 ? " active" : "");
    tab.textContent = name;
    tab.addEventListener("click", function() { switchTab(i); });
    tabBar.appendChild(tab);
    tabEls.push(tab);
  });
  panel.appendChild(tabBar);

  var contents = [];
  TABS.forEach(function(name, i) {
    var c = document.createElement("div");
    c.className = "gd-content";
    c.style.display = i === 0 ? "block" : "none";
    if (i < 2) {
      c.innerHTML = '<div class="gd-empty">No data yet</div>';
    } else if (i === 2) {
      c.innerHTML = '<div class="gd-empty">Waiting for first event...</div>';
    } else {
      c.innerHTML = '<div class="gd-empty">Connecting...</div>';
    }
    panel.appendChild(c);
    contents.push(c);
  });

  shadow.appendChild(panel);

  var visible = false;
  var activeTab = 0;
  var maxEntries = 100;
  var firstEvent = [true, true];

  function switchTab(idx) {
    tabEls[activeTab].classList.remove("active");
    contents[activeTab].style.display = "none";
    activeTab = idx;
    tabEls[activeTab].classList.add("active");
    contents[activeTab].style.display = "block";
  }

  function toggle() {
    visible = !visible;
    panel.classList.toggle("gd-collapsed", !visible);
  }

  toggleBtn.addEventListener("click", toggle);
  header.addEventListener("click", toggle);

  document.addEventListener("keydown", function(e) {
    if (e.ctrlKey && e.shiftKey && e.key === "D") {
      e.preventDefault();
      toggle();
    }
  });

  function ts(unixMs) {
    var d = new Date(unixMs);
    var h = String(d.getHours()).padStart(2, "0");
    var m = String(d.getMinutes()).padStart(2, "0");
    var s = String(d.getSeconds()).padStart(2, "0");
    var ms = String(d.getMilliseconds()).padStart(3, "0");
    return h + ":" + m + ":" + s + "." + ms;
  }

  function esc(s) {
    var div = document.createElement("div");
    div.textContent = s;
    return div.innerHTML;
  }

  function addEntry(container, html, tabIdx) {
    if (firstEvent[tabIdx]) {
      container.innerHTML = "";
      firstEvent[tabIdx] = false;
    }
    var div = document.createElement("div");
    div.className = "gd-entry";
    div.innerHTML = html;
    container.insertBefore(div, container.firstChild);
    while (container.children.length > maxEntries) {
      container.removeChild(container.lastChild);
    }
  }

  window.__gerberaDebug = function(meta) {
    var t = ts(meta.timestamp);

    addEntry(contents[0],
      '<span class="gd-time">' + t + '</span> ' +
      '<span class="gd-event-name">' + esc(meta.event) + '</span> ' +
      '<span class="gd-payload">' + esc(JSON.stringify(meta.payload)) + '</span>',
      0
    );

    addEntry(contents[1],
      '<span class="gd-time">' + t + '</span> ' +
      '<span class="gd-event-name">' + esc(meta.event) + '</span> ' +
      '<span class="gd-badge gd-badge-ok">' + meta.patchCount + ' patches</span> ' +
      '<span class="gd-badge gd-badge-warn">' + meta.durationMs + 'ms</span>',
      1
    );

    var stateStr = "{}";
    try {
      if (typeof meta.viewState === "string") {
        stateStr = JSON.stringify(JSON.parse(meta.viewState), null, 2);
      } else {
        stateStr = JSON.stringify(meta.viewState, null, 2);
      }
    } catch(e) {
      stateStr = String(meta.viewState);
    }
    contents[2].innerHTML = '<div class="gd-json">' + esc(stateStr) + '</div>';

    contents[3].innerHTML =
      '<div class="gd-session-item"><span class="gd-session-label">Session ID</span>' +
      '<span class="gd-session-value">' + esc(meta.sessionId) + '</span></div>' +
      '<div class="gd-session-item"><span class="gd-session-label">TTL</span>' +
      '<span class="gd-session-value">' + esc(meta.sessionTtl) + '</span></div>' +
      '<div class="gd-session-item"><span class="gd-session-label">Status</span>' +
      '<span class="gd-session-value"><span class="gd-status gd-status-connected"></span>Connected</span></div>' +
      '<div class="gd-session-item"><span class="gd-session-label">Last update</span>' +
      '<span class="gd-session-value">' + t + '</span></div>';
  };

  window.__gerberaDebugDisconnect = function() {
    var dots = shadow.querySelectorAll(".gd-status");
    dots.forEach(function(d) {
      d.classList.remove("gd-status-connected");
      d.classList.add("gd-status-disconnected");
    });
    var sessionTab = contents[3];
    var statusItem = sessionTab.querySelector(".gd-session-value .gd-status");
    if (statusItem && statusItem.parentNode) {
      statusItem.parentNode.innerHTML =
        '<span class="gd-status gd-status-disconnected"></span>Disconnected';
    }
  };
})();
