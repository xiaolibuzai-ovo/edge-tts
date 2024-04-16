package edge_tts

import (
	"fmt"
	"golang.org/x/net/context"
	"testing"
)

func Test1(t *testing.T) {
	fmt.Println(1111)
	tts := NewEdgeTTS()
	if tts == nil {
		fmt.Println(111)
		return
	}
	tts.TextToSpeech(context.Background(), "hello,world,你好", "")
}
