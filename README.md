# edge-tts

Use Microsoft Edge's online text-to-speech service from go WITHOUT needing Microsoft Edge or Windows or an API key

# Getting started

```go
go get -u github.com/xiaolibuzai-ovo/edge-tts
```

## Running

### Text-to-Speech

```go
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
```

### voice-list

```go
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

```

## refer
+ [edge_tts_python](https://github.com/rany2/edge-tts)
+ [edge tts api](https://gist.github.com/czyt/a2d83de838c9b65ab14fc18136f53bc6)
