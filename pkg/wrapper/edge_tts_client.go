package edge_tts

import (
	"fmt"
	edge_tts "github.com/xiaolibuzai-ovo/edge-tts/pkg/edge_tts"
	"golang.org/x/net/context"
	"io"
	"os"
)

type EdgeTTS interface {
	TextToSpeech(ctx context.Context, text string, voice string) (io.Reader, error)
	VoiceList()
}

type ttsClient struct {
	communicate *edge_tts.Communicate
}

func NewEdgeTTS() EdgeTTS {
	return &ttsClient{communicate: edge_tts.NewCommunicate()}
}

func (t *ttsClient) TextToSpeech(ctx context.Context, text string, voice string) (io.Reader, error) {
	audio, err := os.OpenFile("test.mp3", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	if t.communicate == nil {
		fmt.Println("nil")
	}
	if t.communicate.SetText(text) == nil {
		fmt.Println("!!!nil")
	}
	if t.communicate.SetText(text).SetVoice(voice) == nil {
		fmt.Println("!!!!!!!!!nil")
	}
	err = t.communicate.SetText(text).SetVoice(voice).WriteStreamTo(audio)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (t ttsClient) VoiceList() {

	return
}
