(() => {
  const root = document.documentElement;
  const themeButton = document.getElementById("theme-toggle");
  const saved = (() => { try { return localStorage.getItem("gosqlite-theme"); } catch (_) { return null; } })();
  if (saved === "dark") root.dataset.theme = "dark";
  themeButton?.addEventListener("click", () => {
    const dark = root.dataset.theme !== "dark";
    if (dark) root.dataset.theme = "dark"; else delete root.dataset.theme;
    try { localStorage.setItem("gosqlite-theme", dark ? "dark" : "light"); } catch (_) {}
  });
  const copy = async (text, button, original) => {
    try {
      await navigator.clipboard.writeText(text);
      button.textContent = "Copied!";
      setTimeout(() => { button.textContent = original; }, 1400);
    } catch (_) {
      button.textContent = "Select and copy";
      setTimeout(() => { button.textContent = original; }, 1800);
    }
  };
  const command = "go get github.com/Aswanidev-vs/gosqlite-ffi";
  const commandButton = document.getElementById("copy-command");
  commandButton?.addEventListener("click", () => copy(command, commandButton, "Copy"));
  const exampleButton = document.getElementById("copy-example");
  const example = document.querySelector(".code-card code")?.innerText || "";
  exampleButton?.addEventListener("click", () => copy(example, exampleButton, "Copy code"));
  document.querySelectorAll(".copy-code").forEach((button) => {
    const code = button.closest(".code-card")?.querySelector("pre code")?.innerText || "";
    button.addEventListener("click", () => copy(code, button, "Copy code"));
  });
})(); 
