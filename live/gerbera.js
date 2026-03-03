(function() {
  var sid = document.documentElement.getAttribute("gerbera-session");
  var csrfMeta = document.querySelector('meta[name="gerbera-csrf"]');
  var csrf = csrfMeta ? csrfMeta.getAttribute("content") : "";
  var proto = location.protocol === "https:" ? "wss:" : "ws:";
  var wsUrl = proto + "//" + location.host + location.pathname + "?gerbera-ws=1&session=" + sid + "&csrf=" + csrf;
  var ws;
  var reconnectAttempts = 0;
  var maxReconnectDelay = 30000;
  var reconnectOverlay = null;

  // Decode HTML entities produced by Go's html.EscapeString.
  // Uses a textarea element so that the browser's HTML parser decodes entities
  // without executing any embedded scripts (textarea content model is text).
  var _decEl;
  function decodeAttr(s) {
    if (s.indexOf("&") === -1) return s;
    if (!_decEl) _decEl = document.createElement("textarea");
    _decEl.innerHTML = s;
    return _decEl.value;
  }

  var SVG_NS = "http://www.w3.org/2000/svg";

  // Check if a node is in the SVG namespace.
  function isSVG(n) {
    return n && n.namespaceURI === SVG_NS;
  }

  // Parse an HTML fragment into a DocumentFragment.
  // When the context node is inside an SVG, parse as SVG XML so that
  // child elements are created in the SVG namespace (not XHTML).
  function parseFragment(html, ctx) {
    if (isSVG(ctx)) {
      var wrap = '<svg xmlns="' + SVG_NS + '">' + html + '</svg>';
      var doc = new DOMParser().parseFromString(wrap, "image/svg+xml");
      if (!doc.querySelector("parsererror")) {
        var frag = document.createDocumentFragment();
        var root = doc.documentElement;
        for (var i = 0; i < root.childNodes.length; i++) {
          frag.appendChild(document.importNode(root.childNodes[i], true));
        }
        return frag;
      }
      // Fallback to HTML parsing on XML parse error
    }
    var t = document.createElement("template");
    t.innerHTML = html;
    return t.content;
  }

  // Track IME composition state to avoid disrupting input during composition.
  var _composing = false;
  document.addEventListener("compositionstart", function() { _composing = true; });
  document.addEventListener("compositionend", function(e) {
    _composing = false;
    // In Safari, the final input event fires before compositionend and gets
    // suppressed by the _composing guard. Manually send the composed value
    // so the server state stays in sync.
    var el = e.target;
    if (el && el.getAttribute) {
      var inputEvt = el.getAttribute("gerbera-input");
      if (inputEvt) {
        var p = {value: el.value};
        var gv = el.getAttribute("gerbera-value");
        if (gv) p.value = gv;
        send(inputEvt, p);
      }
    }
  });

  var EVENTS = ["click","input","change","submit","focus","blur","keydown","dblclick","mouseenter","mouseleave"];
  var TOUCH_EVENTS = ["touchstart","touchend","touchmove"];

  function connect() {
    ws = new WebSocket(wsUrl);
    ws.onopen = onOpen;
    ws.onmessage = onMessage;
    ws.onclose = onClose;
  }

  function onOpen() {
    reconnectAttempts = 0;
    hideReconnectOverlay();
    bind();
    mountComponents();
    // Fire mounted hooks
    document.querySelectorAll("[gerbera-hook]").forEach(function(el) {
      if (el._gbHookMounted) return;
      el._gbHookMounted = true;
      var hookName = el.getAttribute("gerbera-hook");
      if (window.__gerberaHooks && window.__gerberaHooks[hookName]) {
        var hookInstance = window.__gerberaHooks[hookName];
        if (hookInstance.mounted) hookInstance.mounted.call(hookInstance, el);
        el._gbHookInstance = hookInstance;
      }
    });
  }

  function onClose(ev) {
    if (window.__gerberaDebugDisconnect) {
      window.__gerberaDebugDisconnect();
    }
    // Notify hook instances of disconnect
    document.querySelectorAll("[gerbera-hook]").forEach(function(el) {
      if (el._gbHookInstance && el._gbHookInstance.disconnected) {
        el._gbHookInstance.disconnected.call(el._gbHookInstance, el);
      }
    });
    // If WebSocket was rejected (session expired), reload to get a new session
    if (ev && ev.code === 1006 && reconnectAttempts > 2) {
      location.reload();
      return;
    }
    showReconnectOverlay();
    var delay = Math.min(1000 * Math.pow(2, reconnectAttempts), maxReconnectDelay);
    reconnectAttempts++;
    setTimeout(function() { connect(); }, delay);
  }

  function showReconnectOverlay() {
    if (reconnectOverlay) return;
    reconnectOverlay = document.createElement("div");
    reconnectOverlay.id = "gerbera-reconnect-overlay";
    reconnectOverlay.style.cssText = "position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,0.5);z-index:2147483646;display:flex;align-items:center;justify-content:center;";
    var box = document.createElement("div");
    box.style.cssText = "background:#fff;padding:24px 32px;border-radius:8px;font-family:sans-serif;text-align:center;box-shadow:0 4px 24px rgba(0,0,0,0.3);";
    box.innerHTML = '<div style="font-size:18px;margin-bottom:8px;">&#x26A0; Connection Lost</div>' +
      '<div style="color:#666;">Reconnecting...</div>' +
      '<div style="color:#999;font-size:12px;margin-top:4px;">接続が切れました。再接続中...</div>';
    reconnectOverlay.appendChild(box);
    document.body.appendChild(reconnectOverlay);
  }

  function hideReconnectOverlay() {
    if (reconnectOverlay) {
      reconnectOverlay.remove();
      reconnectOverlay = null;
    }
  }

  function send(name, payload) {
    if (ws.readyState === WebSocket.OPEN) {
      // Add loading class
      document.documentElement.classList.add("gerbera-loading");
      ws.send(JSON.stringify({e: name, p: payload}));
    }
  }

  function bind() {
    EVENTS.forEach(function(type) {
      document.querySelectorAll("[gerbera-" + type + "]").forEach(function(el) {
        if (el._gb && el._gb[type]) return;
        if (!el._gb) el._gb = {};
        el._gb[type] = true;
        var handler = function(e) {
          if (type === "submit") e.preventDefault();
          if (type === "keydown" && (e.isComposing || e.keyCode === 229)) return;
          if (type === "input" && _composing) return;
          if (type === "keydown" && (e.key === "ArrowDown" || e.key === "ArrowUp" || e.key === "Enter" || e.key === "Escape")) {
            if (el.getAttribute("role") === "combobox") e.preventDefault();
          }
          var name = el.getAttribute("gerbera-" + type);
          var kf = el.getAttribute("gerbera-key");
          if (kf && e.key !== kf) return;
          var p = {};
          if (type === "input" || type === "change") p.value = el.value;
          if (type === "keydown") p.key = e.key;
          var gv = el.getAttribute("gerbera-value");
          if (gv) p.value = gv;
          if (type === "submit") {
            var form = el.tagName === "FORM" ? el : el.closest("form");
            if (form) {
              new FormData(form).forEach(function(v, k) { p[k] = v; });
            }
          }
          // Check debounce
          var debounceMs = parseInt(el.getAttribute("gerbera-debounce"));
          if (debounceMs > 0) {
            var timerKey = "_gbDebounce_" + type;
            if (el[timerKey]) clearTimeout(el[timerKey]);
            el[timerKey] = setTimeout(function() {
              el[timerKey] = null;
              send(name, p);
            }, debounceMs);
          } else {
            send(name, p);
          }
        };
        el.addEventListener(type, handler);
      });
    });

    // Touch events
    TOUCH_EVENTS.forEach(function(type) {
      document.querySelectorAll("[gerbera-" + type + "]").forEach(function(el) {
        if (el._gbTouch && el._gbTouch[type]) return;
        if (!el._gbTouch) el._gbTouch = {};
        el._gbTouch[type] = true;
        el.addEventListener(type, function(e) {
          var name = el.getAttribute("gerbera-" + type);
          var p = {};
          if (e.touches && e.touches.length > 0) {
            var t = e.touches[0];
            p.clientX = String(t.clientX);
            p.clientY = String(t.clientY);
            p.pageX = String(t.pageX);
            p.pageY = String(t.pageY);
          } else if (e.changedTouches && e.changedTouches.length > 0) {
            var t = e.changedTouches[0];
            p.clientX = String(t.clientX);
            p.clientY = String(t.clientY);
            p.pageX = String(t.pageX);
            p.pageY = String(t.pageY);
          }
          p.touchCount = String(e.touches ? e.touches.length : 0);
          send(name, p);
        });
      });
    });

    // Scroll events with throttle
    document.querySelectorAll("[gerbera-scroll]").forEach(function(el) {
      if (el._gbScroll) return;
      el._gbScroll = true;
      var ms = parseInt(el.getAttribute("gerbera-throttle")) || 100;
      var timer = null;
      el.addEventListener("scroll", function() {
        if (timer) return;
        timer = setTimeout(function() {
          timer = null;
          var name = el.getAttribute("gerbera-scroll");
          send(name, {
            scrollTop: String(el.scrollTop),
            scrollHeight: String(el.scrollHeight),
            clientHeight: String(el.clientHeight),
            scrollLeft: String(el.scrollLeft),
            scrollWidth: String(el.scrollWidth),
            clientWidth: String(el.clientWidth)
          });
        }, ms);
      });
    });

    // Lifecycle hooks - bind new hook elements
    document.querySelectorAll("[gerbera-hook]").forEach(function(el) {
      if (el._gbHookMounted) return;
      el._gbHookMounted = true;
      var hookName = el.getAttribute("gerbera-hook");
      if (window.__gerberaHooks && window.__gerberaHooks[hookName]) {
        var hookInstance = Object.create(window.__gerberaHooks[hookName]);
        if (hookInstance.mounted) hookInstance.mounted.call(hookInstance, el);
        el._gbHookInstance = hookInstance;
      }
    });

    // File upload inputs
    document.querySelectorAll("[gerbera-upload]").forEach(function(el) {
      if (el._gbUpload) return;
      el._gbUpload = true;
      el.addEventListener("change", function() {
        if (!el.files || !el.files.length) return;
        var event = el.getAttribute("gerbera-upload");
        var maxSize = parseInt(el.getAttribute("gerbera-upload-max")) || (10 * 1024 * 1024);
        var fd = new FormData();
        for (var i = 0; i < el.files.length; i++) {
          if (el.files[i].size > maxSize) {
            console.warn("Gerbera: file too large:", el.files[i].name, el.files[i].size);
            continue;
          }
          fd.append("files", el.files[i]);
        }
        var url = location.pathname + "?gerbera-upload=1&session=" + sid + "&csrf=" + csrf + "&event=" + encodeURIComponent(event);
        fetch(url, {method: "POST", body: fd}).then(function(res) {
          if (res.ok) send("gerbera:upload_complete", {event: event});
        });
      });
    });

    // Live navigation links
    document.querySelectorAll("[gerbera-live-link]").forEach(function(el) {
      if (el._gbLiveLink) return;
      el._gbLiveLink = true;
      el.addEventListener("click", function(e) {
        e.preventDefault();
        var href = el.getAttribute("href") || el.getAttribute("gerbera-live-link");
        if (href) {
          send("gerbera:navigate", {url: href});
          history.pushState({}, "", href);
        }
      });
    });
  }

  function resolve(path) {
    var n = document.documentElement;
    for (var i = 0; i < path.length; i++) {
      if (!n.children[path[i]]) return null;
      n = n.children[path[i]];
    }
    return n;
  }

  // Execute JS commands from server
  function executeJSCommands(commands) {
    if (!commands || !commands.length) return;
    commands.forEach(function(cmd) {
      var el = cmd.target ? document.querySelector(cmd.target) : null;
      switch (cmd.cmd) {
        case "scroll_to":
          if (el) {
            var opts = {};
            if (cmd.args && cmd.args.top) opts.top = parseFloat(cmd.args.top);
            if (cmd.args && cmd.args.left) opts.left = parseFloat(cmd.args.left);
            if (cmd.args && cmd.args.behavior) opts.behavior = cmd.args.behavior;
            else opts.behavior = "smooth";
            el.scrollTo(opts);
          }
          break;
        case "scroll_into_pct":
          if (el && cmd.args && cmd.args.pct) {
            var pct = parseFloat(cmd.args.pct);
            el.scrollTop = pct * (el.scrollHeight - el.clientHeight);
          }
          break;
        case "focus":
          if (el) el.focus();
          break;
        case "blur":
          if (el) el.blur();
          break;
        case "set_attr":
          if (el && cmd.args) el.setAttribute(cmd.args.key, cmd.args.value);
          break;
        case "remove_attr":
          if (el && cmd.args) el.removeAttribute(cmd.args.key);
          break;
        case "add_class":
          if (el && cmd.args) el.classList.add(cmd.args["class"]);
          break;
        case "remove_class":
          if (el && cmd.args) el.classList.remove(cmd.args["class"]);
          break;
        case "toggle_class":
          if (el && cmd.args) el.classList.toggle(cmd.args["class"]);
          break;
        case "set_prop":
          if (el && cmd.args) {
            var val = cmd.args.value;
            if (val === "true") val = true;
            else if (val === "false") val = false;
            el[cmd.args.key] = val;
          }
          break;
        case "dispatch":
          if (el && cmd.args) {
            el.dispatchEvent(new Event(cmd.args.event, {bubbles: true}));
          }
          break;
        case "show":
          if (el) el.style.display = "";
          break;
        case "hide":
          if (el) el.style.display = "none";
          break;
        case "toggle":
          if (el) el.style.display = el.style.display === "none" ? "" : "none";
          break;
        case "navigate":
          if (cmd.args && cmd.args.url) location.href = cmd.args.url;
          break;
      }
    });
  }

  function onMessage(ev) {
    // Remove loading class
    document.documentElement.classList.remove("gerbera-loading");

    var data = JSON.parse(ev.data);
    var patches;
    var jsCommands;
    if (Array.isArray(data)) {
      patches = data;
    } else {
      patches = typeof data.patches === "string" ? JSON.parse(data.patches) : (data.patches || []);
      jsCommands = data.js_commands;
      if (data.debug && window.__gerberaDebug) {
        window.__gerberaDebug(data.debug);
      }
    }
    patches.forEach(function(p) {
      var n = resolve(p.path);
      if (!n) return;
      var v = p.val != null ? p.val : "";
      switch (p.op) {
        case "text":
          // If the element has child elements, update only the text node
          // to avoid destroying child elements with textContent.
          if (n.children.length > 0) {
            var tn = null;
            for (var i = 0; i < n.childNodes.length; i++) {
              if (n.childNodes[i].nodeType === 3) { tn = n.childNodes[i]; break; }
            }
            if (tn) { tn.textContent = v; }
            else if (v) { n.insertBefore(document.createTextNode(v), n.firstChild); }
          } else {
            n.textContent = v;
          }
          break;
        case "html":
          if (isSVG(n)) {
            while (n.firstChild) n.removeChild(n.firstChild);
            if (v) n.appendChild(parseFragment(v, n));
          } else {
            n.innerHTML = v;
          }
          break;
        case "attr":
          var dv = decodeAttr(v);
          n.setAttribute(p.key, dv);
          if (p.key === "value" && (n.tagName === "INPUT" || n.tagName === "TEXTAREA" || n.tagName === "SELECT")) {
            if (!(_composing && n === document.activeElement)) {
              n.value = dv;
            }
          }
          break;
        case "rattr":
          n.removeAttribute(p.key);
          break;
        case "class":
          if (isSVG(n)) {
            n.setAttribute("class", v);
          } else {
            n.className = v;
          }
          break;
        case "insert": {
          var frag = parseFragment(p.html, n);
          n.insertBefore(frag, n.children[p.idx] || null);
          break;
        }
        case "remove":
          if (n.children[p.idx]) {
            // Fire destroyed hook
            var removed = n.children[p.idx];
            if (removed._gbHookInstance && removed._gbHookInstance.destroyed) {
              removed._gbHookInstance.destroyed.call(removed._gbHookInstance, removed);
            }
            n.removeChild(removed);
          }
          break;
        case "replace": {
          var frag = parseFragment(p.html, n.parentNode);
          // Fire destroyed hook on old element
          if (n._gbHookInstance && n._gbHookInstance.destroyed) {
            n._gbHookInstance.destroyed.call(n._gbHookInstance, n);
          }
          n.replaceWith(frag);
          break;
        }
      }
    });

    // Always rebind after applying patches — new gerbera-* attributes may
    // have been added via attr/class patches without any insert/replace.
    if (patches.length > 0) bind();

    // Execute JS commands
    executeJSCommands(jsCommands);

    // Fire updated hooks
    document.querySelectorAll("[gerbera-hook]").forEach(function(el) {
      if (el._gbHookInstance && el._gbHookInstance.updated) {
        el._gbHookInstance.updated.call(el._gbHookInstance, el);
      }
    });
  }

  // Hook registry
  window.__gerberaHooks = window.__gerberaHooks || {};
  window.Gerbera = window.Gerbera || {};
  window.Gerbera.registerHook = function(name, hookDef) {
    window.__gerberaHooks[name] = hookDef;
  };

  // Mount sub-components
  function mountComponents() {
    document.querySelectorAll("[gerbera-component]").forEach(function(el) {
      if (el._gbComponentMounted) return;
      el._gbComponentMounted = true;
      var path = el.getAttribute("gerbera-component");
      // Load component via fetch and inject HTML, then connect WebSocket
      fetch(path).then(function(res) { return res.text(); }).then(function(html) {
        // Extract body content from the full HTML
        var parser = new DOMParser();
        var doc = parser.parseFromString(html, "text/html");
        var body = doc.body;
        if (body) {
          el.innerHTML = body.innerHTML;
          // Get session from the component's HTML
          var compSid = doc.documentElement.getAttribute("gerbera-session");
          var compCsrfMeta = doc.querySelector('meta[name="gerbera-csrf"]');
          var compCsrf = compCsrfMeta ? compCsrfMeta.getAttribute("content") : "";
          if (compSid) {
            var compWs = new WebSocket(
              proto + "//" + location.host + path + "?gerbera-ws=1&session=" + compSid + "&csrf=" + compCsrf
            );
            compWs.onmessage = function(ev) {
              // Apply patches scoped to the component container
              var data = JSON.parse(ev.data);
              var patches = Array.isArray(data) ? data : JSON.parse(data.patches || "[]");
              patches.forEach(function(p) {
                // Adjust path resolution to start from the component container
                var n = el;
                for (var i = 0; i < p.path.length; i++) {
                  if (!n.children[p.path[i]]) return;
                  n = n.children[p.path[i]];
                }
                // Apply patch (reuse same logic)
                var v = p.val != null ? p.val : "";
                switch (p.op) {
                  case "text":
                    if (n.children.length > 0) {
                      var tn = null;
                      for (var j = 0; j < n.childNodes.length; j++) {
                        if (n.childNodes[j].nodeType === 3) { tn = n.childNodes[j]; break; }
                      }
                      if (tn) { tn.textContent = v; }
                      else if (v) { n.insertBefore(document.createTextNode(v), n.firstChild); }
                    } else { n.textContent = v; }
                    break;
                  case "html":
                    if (isSVG(n)) {
                      while (n.firstChild) n.removeChild(n.firstChild);
                      if (v) n.appendChild(parseFragment(v, n));
                    } else {
                      n.innerHTML = v;
                    }
                    break;
                  case "attr":
                    var dv = decodeAttr(v);
                    n.setAttribute(p.key, dv);
                    if (p.key === "value" && (n.tagName === "INPUT" || n.tagName === "TEXTAREA" || n.tagName === "SELECT")) {
                      if (!(_composing && n === document.activeElement)) {
                        n.value = dv;
                      }
                    }
                    break;
                  case "rattr": n.removeAttribute(p.key); break;
                  case "class":
                    if (isSVG(n)) {
                      n.setAttribute("class", v);
                    } else {
                      n.className = v;
                    }
                    break;
                  case "insert": {
                    var frag = parseFragment(p.html, n);
                    n.insertBefore(frag, n.children[p.idx] || null);
                    break;
                  }
                  case "remove":
                    if (n.children[p.idx]) n.removeChild(n.children[p.idx]);
                    break;
                  case "replace": {
                    var frag = parseFragment(p.html, n.parentNode);
                    n.replaceWith(frag);
                    break;
                  }
                }
              });
              if (patches.length > 0) bind();
            };
          }
          bind();
        }
      });
    });
  }

  connect();
})();
