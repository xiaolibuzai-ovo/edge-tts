package main

import (
	edgeTTS "github.com/xiaolibuzai-ovo/edge-tts/pkg/wrapper"
	"golang.org/x/net/context"
	"io"
	"os"
)

func main() {
	tts := edgeTTS.NewEdgeTTS()
	voice := "zh-CN-XiaoxiaoNeural"
	speech, err := tts.TextToSpeech(context.Background(), "hello,world", voice)
	if err != nil {
		return
	}

	// speech is io.ReaderCloser
	// such as
	filePath := "./output.mp3"
	file, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	io.Copy(file, speech)
}
