package attribution

import "testing"

// Фиксирует матрицу Phase A классификатора. Пакет наполняет
// `attribution_events`, на которой живёт донат «Code provenance», и с момента,
// когда worker поднимается внутри контейнера api, эта матрица — боевой контракт
// между агентом/расширениями и дашбордом. Категории должны совпадать с Enum8 в
// `migrate/sql/clickhouse/002_attribution_events.up.sql`.
func TestClassify(t *testing.T) {
	cases := []struct {
		name string
		in   rawEvent
		want string
	}{
		{
			name: "IDE inline-accept → ai_inline",
			in:   rawEvent{Source: "ide", Category: "ai", AIChannel: "inline"},
			want: "ai_inline",
		},
		{
			name: "CLI agent с провайдером → ai_agent",
			in:   rawEvent{Source: "cli", AIProvider: "anthropic"},
			want: "ai_agent",
		},
		{
			name: "IDE manual → typed",
			in:   rawEvent{Source: "ide", Category: "manual"},
			want: "typed",
		},
		{
			name: "IDE refactor → refactor",
			in:   rawEvent{Source: "ide", Category: "refactor"},
			want: "refactor",
		},
		{
			name: "browser → unknown (pasted_ai ждёт Phase B)",
			in:   rawEvent{Source: "browser", Category: "ai", AIProvider: "openai"},
			want: "unknown",
		},
		{
			// Регрессия на R3: события десктоп-агента несут keystroke-дельты в
			// chars_in, но Phase A их не различает. Если появится обработка
			// source=os — этот кейс должен осознанно измениться, а не молча.
			name: "OS focus → unknown",
			in:   rawEvent{Source: "os", Category: "manual", CharsIn: 500},
			want: "unknown",
		},
		{
			name: "CLI без провайдера → unknown",
			in:   rawEvent{Source: "cli"},
			want: "unknown",
		},
		{
			name: "IDE ai без inline-канала → unknown",
			in:   rawEvent{Source: "ide", Category: "ai", AIChannel: "chat"},
			want: "unknown",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := classify(tc.in).Category; got != tc.want {
				t.Errorf("classify(%+v).Category = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// Числовые поля должны переезжать в attribution как есть — донат считает
// проценты по focus_ms, а таблицы по lines/chars.
func TestClassifyCarriesMetrics(t *testing.T) {
	got := classify(rawEvent{
		Source: "ide", Category: "manual",
		FileLang: "go", AIProvider: "", DurationMS: 4200, CharsIn: 77, LinesAdded: 9,
	})
	if got.FocusMS != 4200 || got.Chars != 77 || got.Lines != 9 || got.FileLang != "go" {
		t.Errorf("metrics not carried through: %+v", got)
	}
}
