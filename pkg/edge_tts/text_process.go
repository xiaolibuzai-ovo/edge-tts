package edge_tts

import (
	"html"
	"strings"
	"unicode"
)

func escape(data string) string {
	// Must do ampersand first
	entities := make(map[string]string)
	data = html.EscapeString(data)
	data = strings.NewReplacer(">", "&gt;", "<", "&lt;").Replace(data)
	if entities != nil {
		data = replaceWithDict(data, entities)
	}
	return data
}

func replaceWithDict(data string, entities map[string]string) string {
	for key, value := range entities {
		data = strings.ReplaceAll(data, key, value)
	}
	return data
}

func removeIncompatibleCharacters(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return ' '
		}
		return r
	}, str)
}

func calculateMaxMessageSize(pitch, voice string, rate string, volume string) int {
	websocketMaxSize := 1 << 16
	overheadPerMessage := len(ssmlHeadersAppendExtraData(generateConnectID(), currentTimeInMST(), makeSsml("", pitch, voice, rate, volume))) + 50
	return websocketMaxSize - overheadPerMessage
}
