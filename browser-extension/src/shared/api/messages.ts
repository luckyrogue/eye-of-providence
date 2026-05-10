// Discriminated unions для chrome.runtime sendMessage — заменяют untyped any.
// content.ts и popup.tsx импортируют отсюда; background.ts switch'ится по type.

export type AiCopyMessage = {
  type: "ai-copy";
  host: string;
  size: number;
};

export type FlushNowMessage = {
  type: "flush-now";
};

export type PendingCountMessage = {
  type: "pending-count";
};

export type ExtensionMessage = AiCopyMessage | FlushNowMessage | PendingCountMessage;

export type AiCopyResponse = { ok: true };
export type FlushNowResponse = { ok: true };
export type PendingCountResponse = { buffer: number; retry: number };
