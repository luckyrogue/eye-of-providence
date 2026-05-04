// AI domain → provider/channel mapping. Single source of truth для background и content scripts.

export type AiInfo = { provider: string; channel: string };

const MAP: Record<string, AiInfo> = {
  "chatgpt.com": { provider: "openai", channel: "chat" },
  "chat.openai.com": { provider: "openai", channel: "chat" },
  "claude.ai": { provider: "anthropic", channel: "chat" },
  "gemini.google.com": { provider: "google", channel: "chat" },
  "copilot.microsoft.com": { provider: "microsoft", channel: "chat" },
  "www.perplexity.ai": { provider: "perplexity", channel: "chat" },
  "perplexity.ai": { provider: "perplexity", channel: "chat" },
  "poe.com": { provider: "poe", channel: "chat" },
  "you.com": { provider: "you", channel: "chat" },
  "cursor.sh": { provider: "cursor", channel: "agent" },
  "www.phind.com": { provider: "phind", channel: "chat" },
  "bolt.new": { provider: "stackblitz", channel: "agent" },
  "v0.dev": { provider: "vercel", channel: "agent" },
  "lovable.dev": { provider: "lovable", channel: "agent" },
};

export function aiInfoForHost(host: string): AiInfo | null {
  return MAP[host] ?? null;
}

export function isAiHost(host: string): boolean {
  return host in MAP;
}
