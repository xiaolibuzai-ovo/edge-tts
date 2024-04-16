package edge_tts

import (
	"bytes"
	"html"
	"strings"
	"unicode"
)

func splitTextByByteLength(text string, byteLength int) [][]byte {
	var result [][]byte
	textBytes := []byte(text)

	if byteLength > 0 {
		for len(textBytes) > byteLength {
			splitAt := bytes.LastIndexByte(textBytes[:byteLength], ' ')
			if splitAt == -1 || splitAt == 0 {
				splitAt = byteLength
			} else {
				splitAt++
			}

			trimmedText := bytes.TrimSpace(textBytes[:splitAt])
			if len(trimmedText) > 0 {
				result = append(result, trimmedText)
			}
			textBytes = textBytes[splitAt:]
		}
	}

	trimmedText := bytes.TrimSpace(textBytes)
	if len(trimmedText) > 0 {
		result = append(result, trimmedText)
	}

	return result
}

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
