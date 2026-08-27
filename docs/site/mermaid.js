import mermaid from "https://cdn.jsdelivr.net/npm/mermaid@11.12.0/dist/mermaid.esm.min.mjs";

for (const code of document.querySelectorAll("pre > code.language-mermaid")) {
  const diagram = document.createElement("div");
  diagram.className = "mermaid";
  diagram.textContent = code.textContent;
  code.parentElement.replaceWith(diagram);
}

mermaid.initialize({
  startOnLoad: true,
  securityLevel: "strict",
  theme: "base",
  flowchart: { curve: "basis", htmlLabels: true },
  themeVariables: {
    background: "#0b151d",
    primaryColor: "#142832",
    primaryTextColor: "#edf5f7",
    primaryBorderColor: "#57e3ce",
    secondaryColor: "#1a2834",
    secondaryTextColor: "#edf5f7",
    secondaryBorderColor: "#587382",
    tertiaryColor: "#101d27",
    tertiaryTextColor: "#edf5f7",
    tertiaryBorderColor: "#f4bd68",
    lineColor: "#78909c",
    textColor: "#edf5f7",
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif",
    fontSize: "14px"
  }
});
