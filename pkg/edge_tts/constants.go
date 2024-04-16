package edge_tts

const (
	trustedClientToken = "6A5AA1D4EAFF4E9FB37E23D68491D6F4"
	EdgeWssEndpoint    = "wss://speech.platform.bing.com/consumer/speech/synthesize/" + "readaloud/edge/v1?trustedClientToken=" + trustedClientToken
	VoiceListEndpoint  = "https://speech.platform.bing.com/consumer/speech/synthesize/readaloud/voices/list?trustedclienttoken=" + trustedClientToken
)

const (
	PathTurnStart     = "turn.start"
	PathTurnEnd       = "turn.end"
	PathAudioMetadata = "audio.metadata"
	PathResponse      = "response"
)

const (
	DefaultVoice  = "zh-CN-XiaoxiaoNeural"
	DefaultPitch  = "+0Hz"
	DefaultRate   = "+0%"
	DefaultVolume = "+0%"
)
