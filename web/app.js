const form = document.getElementById("run-form");
const cmdlineInput = document.getElementById("cmdline");
const workdirInput = document.getElementById("workdir");
const timeoutInput = document.getElementById("timeout");
const backgroundInput = document.getElementById("background");
const outputEl = document.getElementById("output");
const statusEl = document.getElementById("status");
const clearBtn = document.getElementById("clear");
const bindInfo = document.getElementById("bind-info");

if (bindInfo) {
  bindInfo.textContent = window.location.host || "this host";
}

form.addEventListener("submit", async (event) => {
  event.preventDefault();
  const cmdline = cmdlineInput.value.trim();
  if (!cmdline) {
    setStatus("Please enter a command.");
    return;
  }

  const parsed = parseCmdline(cmdline);
  if (!parsed) {
    setStatus("Could not parse command line.");
    return;
  }

  const payload = { cmd: parsed.cmd, args: parsed.args };
  const wd = workdirInput.value.trim();
  if (wd) payload.workdir = wd;
  const timeout = parseInt(timeoutInput.value, 10);
  if (!Number.isNaN(timeout) && timeout > 0) payload.timeout_sec = timeout;
   if (backgroundInput && backgroundInput.checked) payload.background = true;

  outputEl.textContent = `$ ${cmdline}\n`;
  setStatus("Running...");
  try {
    const resp = await fetch("/api/run", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    if (!resp.ok) {
      const text = await resp.text();
      outputEl.textContent += `Error: ${text}`;
      setStatus(`HTTP ${resp.status}`);
      return;
    }

    const ct = resp.headers.get("content-type") || "";
    if (ct.includes("application/json")) {
      const data = await resp.json();
      outputEl.textContent += `Background started. PID: ${data.pid ?? "?"}\nTimeout: ${data.timeout_sec ?? 0}s\n`;
      setStatus("Started (background)");
      return;
    }

    const reader = resp.body.getReader();
    const decoder = new TextDecoder();
    while (true) {
      const { value, done } = await reader.read();
      if (done) break;
      outputEl.textContent += decoder.decode(value, { stream: true });
      outputEl.scrollTop = outputEl.scrollHeight;
    }
    setStatus("Done");
  } catch (err) {
    outputEl.textContent += `Error: ${err.message}`;
    setStatus("Failed");
  }
});

clearBtn.addEventListener("click", () => {
  outputEl.textContent = "";
  setStatus("");
});

function setStatus(text) {
  statusEl.textContent = text;
}

function parseCmdline(input) {
  const parts = input.match(/"[^"]*"|'[^']*'|[^\s]+/g) || [];
  if (!parts.length) return null;
  const cmd = stripQuotes(parts[0]);
  const args = parts.slice(1).map(stripQuotes);
  return { cmd, args };
}

function stripQuotes(token) {
  if (!token) return token;
  if ((token.startsWith('"') && token.endsWith('"')) || (token.startsWith("'") && token.endsWith("'"))) {
    return token.slice(1, -1);
  }
  return token;
}
