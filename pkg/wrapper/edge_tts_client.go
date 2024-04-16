package edgeTTS

import (
	edge_tts "github.com/xiaolibuzai-ovo/edge-tts/pkg/edge_tts"
	"golang.org/x/net/context"
	"io"
)

type EdgeTTS interface {
	TextToSpeech(ctx context.Context, text string, voice string) (io.ReadCloser, error)
	VoiceList(ctx context.Context) ([]edge_tts.VoiceItem, error)
}

type ttsClient struct {
	communicate *edge_tts.Communicate
}

func NewEdgeTTS() EdgeTTS {
	return &ttsClient{communicate: edge_tts.NewCommunicate()}
}

func (t *ttsClient) TextToSpeech(ctx context.Context, text string, voice string) (io.ReadCloser, error) {
	pr, err := t.communicate.SetText(text).SetVoice(voice).ReadStream()
	if err != nil {
		return nil, err
	}
	return pr, nil
}

func (t *ttsClient) VoiceList(ctx context.Context) ([]edge_tts.VoiceItem, error) {
	list, err := edge_tts.VoiceList()
	if err != nil {
		return nil, err
	}
	return list, nil
}
