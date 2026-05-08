// Ответы /v1/me совпадают с domain types — реэкспортируем как Res-алиасы
// для единообразия (типизация запросов через `http.get<MeRes>(...)`).
import type { Me, Profile } from "./types";

type MeRes = Me;
type ProfileRes = Profile;

export type { MeRes, ProfileRes };
