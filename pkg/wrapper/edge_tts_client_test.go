package edgeTTS

import (
	"fmt"
	"github.com/xiaolibuzai-ovo/edge-tts/pkg/edge_tts"
	"golang.org/x/net/context"
	"io"
	"os"
	"testing"
)

func TestTtsClient_TextToSpeech(t *testing.T) {
	tts := NewEdgeTTS()
	speech, err := tts.TextToSpeech(context.Background(), "hello,world,你好", "")
	if err != nil {
		fmt.Println(err)
		return
	}

	filePath := "./output.mp3"
	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer file.Close()
	_, err = io.Copy(file, speech)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}
	fmt.Printf("Audio stream saved to: %s\n", filePath)
}

func TestTtsClient_VoiceList(t *testing.T) {
	list, err := edge_tts.VoiceList()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(list)
}
