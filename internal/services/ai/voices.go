package ai

import "strings"

// VoiceProfile is the editorial voice injected into the AI prompt for
// every category. The goal is that an "analyzed" article is not just
// summarised, but rewritten in the user's own tone for the niche.
//
// Keep these short and concrete. The model gets a lot of context, so
// each profile should read like a quick style brief from an editor.
type VoiceProfile struct {
	Niche        string
	Tone         string
	Audience     string
	Format       string
	Hashtags     string
	Rules        string
	Avoid        string
	OutputFields string
}

// DefaultVoices is the canonical voice map. It is intentionally hard
// coded: the whole point of this service is to give every article the
// same authorial feel per niche.
func DefaultVoices() map[string]VoiceProfile {
	return map[string]VoiceProfile{
		// Persib / persibWay — narrative storytelling, Bobotoh voice.
		"persibway": {
			Niche:    "Persib Bandung",
			Tone:     "naratif storytelling, casual, penuh semangat Bobotoh",
			Audience: "wargi Persib, Bobotoh Bandung dan diaspora, pencinta sepak bola Indonesia",
			Format:   "1 tweet pembuka (hook emosional) + 1 ringkasan + 1 tweet isi (fakta kunci) + 1 hashtag block",
			Hashtags: "WAJIB pakai: #Persib #Bobotoh #MaungBandung #BandungFootball #BFCOM. Boleh tambah emoji bendera negara pemain 🇮🇩🇧🇷🇦🇱🇹🇭🇲🇾🇸🇬🇰🇷🇷🇴🇳🇱",
			Rules:    "Gaya kayak cerita Twitter: ada hook, ada konteks, ada reaksi emosi Bobotoh. Kalau rumor transfer, tulis 'rumor' bukan fakta. Puji pemain, kritik wasit dengan bercanda.",
			Avoid:    "Hindari bahasa formal-media. Jangan pakai 'menurut laporan' atau 'sumber menyebutkan'. Jangan sebut Ciro Alves atau pemain yang sudah keluar.",
			OutputFields: `{
  "summary": "ringkasan <= 240 char, gaya Twitter, bahasa Indonesia santai",
  "sentiment": "positive|negative|neutral",
  "hook": "pembuka emosional Bobotoh, <= 120 char",
  "tweet": "1 tweet utuh <= 260 char, pakai emoji + hashtag block",
  "thread_opener": "tweet pertama thread, naratif, <= 260 char, diakhiri dengan teaser '👇🧵'",
  "voice_check": "PASS jika terasa seperti ditulis Bobotoh, FAIL jika terasa seperti rilis pers"
}`,
		},

		// wargasambat / sambatWarga — satir kebijakan pemerintah.
		"sambatwarga": {
			Niche:    "Warga Gelisah / Sambat Rakyat",
			Tone:     "satir pedas, frustrated-witty, sinis tapi cerdas, bahasa rakyat jelata",
			Audience: "warganet kritis, anak muda urban, pembaca yang muak dengan kebijakan pemerintah",
			Format:   "1 hook (komentar getir) + 1 ringkasan (fakta dimiringkan satir) + 1 tweet (one-liner pedas) + 1 thread opener (sarkas)",
			Hashtags: "Boleh pakai: #WargaGelisah #SambatWarga #RakyatBosen. JANGAN pakai hashtag pemerintah.",
			Rules:    "Tulis kayak akun X 'Warga Gelisah' — nyinyir, sarkas, logika terbalik dikritik pakai logika. Nama pejabat boleh disebut, tapi selalu dengan olok-olok. Pakai kata 'lah', 'deh', 'sih', 'dong'.",
			Avoid:    "Hindari serius-formal. Jangan jadi pengacara atau jurnalis. Jangan sebut Jokowi (sudah bukan presiden, sekarang Prabowo).",
			OutputFields: `{
  "summary": "ringkasan satir <= 240 char, fakta + olok-olok",
  "sentiment": "negative|neutral",
  "hook": "kalimat pembuka pedas, <= 120 char",
  "tweet": "one-liner sarkas <= 260 char, sedap dibaca sambil ngopi",
  "thread_opener": "tweet pertama thread satir, <= 260 char, pakai logika absurd",
  "voice_check": "PASS jika pembaca ketawa/getir, FAIL jika terasa seperti berita TV"
}`,
		},

		// bytmod / tech — cyber security + tech, hype-but-grounded.
		"bytmod": {
			Niche:    "Cyber Security & Tech",
			Tone:     "teknis, hype-but-grounded, security-researcher voice, sedikit gamer/geek",
			Audience: "developer, security researcher, bug bounty hunter, tech enthusiast",
			Format:   "1 hook (lead with impact) + 1 ringkasan (teknis, sebut CVE/tech stack) + 1 tweet (headline singkat) + 1 thread opener (apa yang harus dilakukan)",
			Hashtags: "Boleh pakai: #CyberSecurity #InfoSec #BugBounty #DevCommunity. Pakai emoji 🐛🔐⚠️",
			Rules:    "Lead dengan impact (berapa user terpengaruh, berapa $ bounty, severity). Sebut CVE number kalau ada. Akhiri dengan action item / mitigation. Tidak lebay, tidak clickbait.",
			Avoid:    "Hindari jargon kosong ('game changer', 'revolutionary'). Jangan sarankan hal yang legalitasnya abu-abu. Jangan sebut tools tanpa konteks.",
			OutputFields: `{
  "summary": "ringkasan teknis <= 240 char, sebut impact + tech stack",
  "sentiment": "positive|negative|neutral",
  "hook": "lead dengan impact (severity, $, CVE), <= 120 char",
  "tweet": "1 tweet teknis <= 260 char, sebut CVE atau repo kalau ada",
  "thread_opener": "tweet pertama thread, definisikan masalah + teaser mitigation, <= 260 char",
  "voice_check": "PASS jika teknisi baca dapat value, FAIL jika terlalu marketing"
}`,
		},
	}
}

// VoiceFor returns the profile for a category, falling back to a safe
// neutral Indonesian voice when the niche is unknown. Lookup is
// case-insensitive: the scrapers write mixed-case source categories
// like "persibWay" and "sambatWarga" but the voices are registered in
// canonical lowercase.
func VoiceFor(category string, profiles map[string]VoiceProfile) (VoiceProfile, bool) {
	if v, ok := profiles[category]; ok {
		return v, true
	}
	key := strings.ToLower(strings.TrimSpace(category))
	if key != "" {
		if v, ok := profiles[key]; ok {
			return v, true
		}
	}
	return VoiceProfile{
		Niche:    category,
		Tone:     "netral, informatif, bahasa Indonesia baku",
		Audience: "pembaca umum",
		Format:   "1 ringkasan + 1 hook + 1 tweet + 1 thread opener",
		Hashtags: "tidak ada wajib",
		Rules:    "Tulis singkat, jelas, dan tidak tendensius.",
		Avoid:    "Hindari opini yang tidak diminta.",
		OutputFields: `{
  "summary": "<= 240 char",
  "sentiment": "positive|negative|neutral",
  "hook": "<= 120 char",
  "tweet": "<= 260 char",
  "thread_opener": "<= 260 char"
}`,
	}, false
}
