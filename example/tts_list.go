package main

import (
	"fmt"
	edgeTTS "github.com/xiaolibuzai-ovo/edge-tts/pkg/wrapper"
	"golang.org/x/net/context"
)

func main() {
	tts := edgeTTS.NewEdgeTTS()
	voiceList, err := tts.VoiceList(context.Background())
	if err != nil {
		return
	}

	fmt.Println(voiceList)
}
