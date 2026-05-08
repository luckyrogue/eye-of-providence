import type { EventRow } from "./types";

type IngestReq = { events: Partial<EventRow & { file_lang: string }>[] };

export type { IngestReq };
