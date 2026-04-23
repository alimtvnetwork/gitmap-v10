import { useState, useCallback, useMemo } from "react";
import { Copy, Check, Download, Maximize2, Minimize2, AArrowUp, AArrowDown } from "lucide-react";
import { copyToClipboard } from "@/lib/clipboard";
import hljs from "highlight.js/lib/core";
import go from "highlight.js/lib/languages/go";
import typescript from "highlight.js/lib/languages/typescript";
import javascript from "highlight.js/lib/languages/javascript";
import bash from "highlight.js/lib/languages/bash";
import json from "highlight.js/lib/languages/json";
import sql from "highlight.js/lib/languages/sql";
import css from "highlight.js/lib/languages/css";
import xml from "highlight.js/lib/languages/xml";
import yaml from "highlight.js/lib/languages/yaml";
import markdown from "highlight.js/lib/languages/markdown";
import powershell from "highlight.js/lib/languages/powershell";
import rust from "highlight.js/lib/languages/rust";
import php from "highlight.js/lib/languages/php";
import cpp from "highlight.js/lib/languages/cpp";
import csharp from "highlight.js/lib/languages/csharp";

hljs.registerLanguage("go", go);
hljs.registerLanguage("typescript", typescript);
hljs.registerLanguage("ts", typescript);
hljs.registerLanguage("javascript", javascript);
hljs.registerLanguage("js", javascript);
hljs.registerLanguage("bash", bash);
hljs.registerLanguage("shell", bash);
hljs.registerLanguage("sh", bash);
hljs.registerLanguage("json", json);
hljs.registerLanguage("sql", sql);
hljs.registerLanguage("css", css);
hljs.registerLanguage("html", xml);
hljs.registerLanguage("xml", xml);
hljs.registerLanguage("yaml", yaml);
hljs.registerLanguage("yml", yaml);
hljs.registerLanguage("markdown", markdown);
hljs.registerLanguage("md", markdown);
hljs.registerLanguage("powershell", powershell);
hljs.registerLanguage("ps1", powershell);
hljs.registerLanguage("rust", rust);
hljs.registerLanguage("php", php);
hljs.registerLanguage("cpp", cpp);
hljs.registerLanguage("csharp", csharp);

interface CodeBlockProps {
  code: string;
  language?: string;
  title?: string;
}

const LANG_COLORS: Record<string, string> = {
  typescript: "99 83% 62%",
  ts: "99 83% 62%",
  javascript: "53 93% 54%",
  js: "53 93% 54%",
  go: "194 66% 55%",
  php: "234 45% 60%",
  css: "264 55% 58%",
  json: "38 92% 50%",
  bash: "120 40% 55%",
  shell: "120 40% 55%",
  sh: "120 40% 55%",
  sql: "200 70% 55%",
  rust: "25 85% 55%",
  html: "12 80% 55%",
  xml: "12 80% 55%",
  yaml: "0 75% 55%",
  yml: "0 75% 55%",
  markdown: "252 85% 60%",
  md: "252 85% 60%",
  powershell: "210 60% 55%",
  ps1: "210 60% 55%",
  cpp: "200 50% 55%",
  csharp: "270 60% 55%",
};

const LANG_EXTENSIONS: Record<string, string> = {
  typescript: "ts", ts: "ts", javascript: "js", js: "js",
  go: "go", php: "php", css: "css", json: "json",
  bash: "sh", shell: "sh", sh: "sh", sql: "sql",
  rust: "rs", html: "html", xml: "xml", yaml: "yml",
  yml: "yml", markdown: "md", md: "md", powershell: "ps1", ps1: "ps1",
  cpp: "cpp", csharp: "cs",
};

const DEFAULT_ACCENT = "220 10% 50%";

const FONT_SIZES = [
  { label: "S", size: "13px" },
  { label: "M", size: "15px" },
  { label: "L", size: "17px" },
];

const CodeBlock = ({ code, language = "bash", title }: CodeBlockProps) => {
  const [copied, setCopied] = useState(false);
  const [fullscreen, setFullscreen] = useState(false);
  const [pinnedLines, setPinnedLines] = useState<Set<number>>(new Set());
  const [lastPinned, setLastPinned] = useState<number | null>(null);
  const [fontSizeIdx, setFontSizeIdx] = useState(1); // default Medium

  const hasPinned = pinnedLines.size > 0;
  const fontSize = FONT_SIZES[fontSizeIdx].size;

  const cycleFontSize = useCallback((direction: "up" | "down") => {
    setFontSizeIdx((prev) => {
      if (direction === "up") return Math.min(prev + 1, FONT_SIZES.length - 1);
      return Math.max(prev - 1, 0);
    });
  }, []);

  const togglePin = useCallback((lineIndex: number, e?: React.MouseEvent) => {
    if (e?.shiftKey && lastPinned !== null) {
      const start = Math.min(lastPinned, lineIndex);
      const end = Math.max(lastPinned, lineIndex);
      setPinnedLines((prev) => {
        const next = new Set(prev);
        for (let i = start; i <= end; i++) next.add(i);
        return next;
      });
    } else {
      setPinnedLines((prev) => {
        const next = new Set(prev);
        if (next.has(lineIndex)) next.delete(lineIndex);
        else next.add(lineIndex);
        return next;
      });
    }
    setLastPinned(lineIndex);
  }, [lastPinned]);

  const handleCopy = useCallback(async () => {
    const textToCopy = hasPinned
      ? Array.from(pinnedLines)
          .sort((a, b) => a - b)
          .map((i) => code.split("\n")[i])
          .join("\n")
      : code;
    await copyToClipboard(textToCopy);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }, [code, hasPinned, pinnedLines]);

  const handleDownload = useCallback(() => {
    const ext = LANG_EXTENSIONS[language.toLowerCase()] ?? "txt";
    const blob = new Blob([code], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `snippet.${ext}`;
    a.click();
    URL.revokeObjectURL(url);
  }, [code, language]);

  const lines = useMemo(() => code.split("\n"), [code]);
  const accent = LANG_COLORS[language.toLowerCase()] ?? DEFAULT_ACCENT;
  const label = language.toUpperCase();
  const showLineNumbers = lines.length > 1;

  const highlightedLines = useMemo(() => {
    const lang = language.toLowerCase();
    let html: string | null = null;
    try {
      if (hljs.getLanguage(lang)) {
        html = hljs.highlight(code, { language: lang }).value;
      }
    } catch {
      // fall through
    }
    if (html) {
      // Split highlighted HTML by newlines, preserving open spans across lines
      const result: string[] = [];
      let openSpans: string[] = [];
      const rawLines = html.split("\n");
      for (const line of rawLines) {
        // Prepend any spans that were open from previous lines
        const prefix = openSpans.join("");
        const full = prefix + line;
        // Track open/close spans
        const opens = line.match(/<span[^>]*>/g) || [];
        const closes = line.match(/<\/span>/g) || [];
        // Update stack
        for (const o of opens) openSpans.push(o);
        for (let i = 0; i < closes.length; i++) openSpans.pop();
        // Close any still-open spans for this line's HTML
        const suffix = "</span>".repeat(openSpans.length);
        result.push(full + suffix);
      }
      return result;
    }
    return null;
  }, [code, language]);

  const wrapperClass = fullscreen
    ? "fixed inset-8 z-[999] rounded-xl flex flex-col"
    : "rounded-xl overflow-hidden my-4";

  return (
    <>
      {fullscreen && (
        <div
          className="fixed inset-0 z-[998] bg-black/70 backdrop-blur-sm"
          onClick={() => setFullscreen(false)}
        />
      )}
      <div
        className={`${wrapperClass} border border-[hsl(220,13%,22%)] group transition-all duration-300`}
        style={{
          ["--lang-accent" as string]: accent,
          background: "hsl(220, 14%, 11%)",
          boxShadow: fullscreen
            ? `0 25px 80px hsl(${accent} / 0.25), 0 0 0 1px hsl(${accent} / 0.3)`
            : undefined,
        }}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-4 py-2 border-b border-[hsl(220,13%,20%)]" style={{ background: "hsl(220, 14%, 14%)" }}>
          <div className="flex items-center gap-2">
            <span
              className="w-[7px] h-[7px] rounded-full"
              style={{
                background: `hsl(${accent})`,
                boxShadow: `0 0 6px hsl(${accent})`,
              }}
            />
            <span className="text-xs font-mono font-medium" style={{ color: `hsl(${accent})` }}>
              {label}
            </span>
            {title && (
              <span className="text-xs font-mono text-[hsl(220,10%,50%)] ml-2">— {title}</span>
            )}
            <span className="text-xs font-mono text-[hsl(220,10%,40%)] ml-2">
              {hasPinned
                ? `${pinnedLines.size} selected`
                : `${lines.length} ${lines.length === 1 ? "line" : "lines"}`}
            </span>
            {hasPinned && (
              <button
                onClick={() => setPinnedLines(new Set())}
                className="text-xs font-mono ml-2 px-1.5 py-0.5 rounded hover:bg-white/10 transition-colors"
                style={{ color: `hsl(${accent})` }}
              >
                Clear
              </button>
            )}
          </div>
          <div className="flex items-center gap-1">
            <button
              onClick={() => cycleFontSize("down")}
              className="p-1.5 rounded hover:bg-white/10 text-[hsl(220,10%,50%)] hover:text-white transition-colors"
              title="Decrease font size"
            >
              <AArrowDown className="h-3.5 w-3.5" />
            </button>
            <span className="text-[10px] font-mono text-[hsl(220,10%,45%)] min-w-[14px] text-center">
              {FONT_SIZES[fontSizeIdx].label}
            </span>
            <button
              onClick={() => cycleFontSize("up")}
              className="p-1.5 rounded hover:bg-white/10 text-[hsl(220,10%,50%)] hover:text-white transition-colors"
              title="Increase font size"
            >
              <AArrowUp className="h-3.5 w-3.5" />
            </button>
            <div className="w-px h-4 bg-[hsl(220,10%,25%)] mx-0.5" />
            <button
              onClick={handleCopy}
              className="p-1.5 rounded hover:bg-white/10 text-[hsl(220,10%,50%)] hover:text-white transition-colors"
              title="Copy"
            >
              {copied ? <Check className="h-3.5 w-3.5 text-primary" /> : <Copy className="h-3.5 w-3.5" />}
            </button>
            <button
              onClick={handleDownload}
              className="p-1.5 rounded hover:bg-white/10 text-[hsl(220,10%,50%)] hover:text-white transition-colors"
              title="Download"
            >
              <Download className="h-3.5 w-3.5" />
            </button>
            <button
              onClick={() => setFullscreen(!fullscreen)}
              className="p-1.5 rounded hover:bg-white/10 text-[hsl(220,10%,50%)] hover:text-white transition-colors"
              title={fullscreen ? "Exit fullscreen" : "Fullscreen"}
            >
              {fullscreen ? <Minimize2 className="h-3.5 w-3.5" /> : <Maximize2 className="h-3.5 w-3.5" />}
            </button>
          </div>
        </div>

        {/* Body */}
        <div className={`overflow-auto ${fullscreen ? "flex-1" : "max-h-[500px]"}`}>
          <div className="flex">
            {showLineNumbers && (
              <div
                className="flex flex-col text-right select-none px-3 py-4 text-xs font-mono border-r border-[hsl(220,13%,18%)] code-line-numbers"
                style={{ background: "hsl(220, 14%, 9%)", color: "hsl(220, 10%, 35%)" }}
              >
                {lines.map((_, i) => (
                  <span
                    key={i}
                    className={`leading-relaxed code-line-num cursor-pointer ${pinnedLines.has(i) ? "code-line-num-pinned" : ""}`}
                    data-line={i}
                    onClick={(e) => togglePin(i, e)}
                  >
                    {i + 1}
                  </span>
                ))}
              </div>
            )}
            <pre className="flex-1 overflow-x-auto leading-relaxed m-0 py-4" style={{ fontSize }}>
              <code className="font-mono hljs block">
                {highlightedLines ? (
                  highlightedLines.map((lineHtml, i) => (
                    <span
                      key={i}
                      className={`code-line block px-4 cursor-pointer ${pinnedLines.has(i) ? "code-line-pinned" : ""}`}
                      onClick={(e) => togglePin(i, e)}
                      dangerouslySetInnerHTML={{ __html: lineHtml || "\n" }}
                    />
                  ))
                ) : (
                  lines.map((line, i) => (
                    <span
                      key={i}
                      className={`code-line block px-4 cursor-pointer ${pinnedLines.has(i) ? "code-line-pinned" : ""}`}
                      onClick={(e) => togglePin(i, e)}
                      style={{ color: "hsl(220, 20%, 92%)" }}
                    >
                      {line || "\n"}
                    </span>
                  ))
                )}
              </code>
            </pre>
          </div>
        </div>
      </div>
    </>
  );
};

export default CodeBlock;
