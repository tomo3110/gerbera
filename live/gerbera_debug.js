(function() {
  var host = document.createElement("div");
  host.id = "gerbera-debug-host";
  host.style.cssText = "position:fixed;bottom:0;right:0;z-index:2147483647;";
  document.body.appendChild(host);
  var shadow = host.attachShadow({mode:"closed"});

  shadow.innerHTML = /*__GERBERA_DEBUG_HTML__*/"";

  var toggleBtn = shadow.querySelector(".gd-toggle");
  var panel = shadow.querySelector(".gd-panel");
  var header = shadow.querySelector(".gd-header");
  var headerStatus = header.querySelector(".gd-status");
  var tabEls = Array.from(shadow.querySelectorAll(".gd-tab"));
  var contents = Array.from(shadow.querySelectorAll(".gd-content"));

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

  tabEls.forEach(function(tab, i) {
    tab.addEventListener("click", function() { switchTab(i); });
  });
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
