(function() {
  var sid = document.documentElement.getAttribute("gerbera-session");
  var proto = location.protocol === "https:" ? "wss:" : "ws:";
  var ws = new WebSocket(
    proto + "//" + location.host + location.pathname + "?gerbera-ws=1&session=" + sid
  );

  var EVENTS = ["click","input","change","submit","focus","blur","keydown"];

  function bind() {
    EVENTS.forEach(function(type) {
      document.querySelectorAll("[gerbera-" + type + "]").forEach(function(el) {
        if (el._gb) return;
        el._gb = true;
        el.addEventListener(type, function(e) {
          if (type === "submit") e.preventDefault();
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
          ws.send(JSON.stringify({e: name, p: p}));
        });
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

  ws.onmessage = function(ev) {
    JSON.parse(ev.data).forEach(function(p) {
      var n = resolve(p.path);
      if (!n) return;
      switch (p.op) {
        case "text":
          n.textContent = p.val;
          break;
        case "attr":
          n.setAttribute(p.key, p.val);
          break;
        case "rattr":
          n.removeAttribute(p.key);
          break;
        case "class":
          n.className = p.val;
          break;
        case "insert": {
          var t = document.createElement("template");
          t.innerHTML = p.html;
          n.insertBefore(t.content, n.children[p.idx] || null);
          bind();
          break;
        }
        case "remove":
          if (n.children[p.idx]) n.removeChild(n.children[p.idx]);
          break;
        case "replace": {
          var t = document.createElement("template");
          t.innerHTML = p.html;
          n.replaceWith(t.content);
          bind();
          break;
        }
      }
    });
  };

  ws.onclose = function() {
    setTimeout(function() { location.reload(); }, 3000);
  };

  bind();
})();
