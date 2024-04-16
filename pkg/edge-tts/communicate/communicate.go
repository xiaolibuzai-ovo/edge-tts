package communicate

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"html"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
	"unicode"
	"wsDemo/aaaa/constant"
)

var (
	escapeReplacer = strings.NewReplacer(">", "&gt;", "<", "&lt;")
)

type Communicate struct {
	text string

	audioDataIndex int
	prevIdx        int
	shiftTime      int
	finalUtterance map[int]int
	op             chan map[string]interface{}
	opt            *CommunicateOption
}

type CommunicateOption struct {
	Voice  string
	Pitch  string
	Rate   string
	Volume string

	HttpProxy        string
	Socket5Proxy     string
	Socket5ProxyUser string
	Socket5ProxyPass string
	IgnoreSSL        bool
}

func NewCommunicate(text string, opt *CommunicateOption) (*Communicate, error) {
	if opt == nil {
		opt = &CommunicateOption{}
	}
	opt.CheckAndApplyDefaultOption()

	if err := WithCommunicateOption(opt); err != nil {
		return nil, err
	}
	return &Communicate{
		text: text,
		opt:  opt,
	}, nil
}

func (c *CommunicateOption) CheckAndApplyDefaultOption() {
	// Default values
	if c.Voice == "" {
		c.Voice = constant.DefaultVoice
	}
	if c.Pitch == "" {
		c.Pitch = "+0Hz"
	}
	if c.Rate == "" {
		c.Rate = "+0%"
	}
	if c.Volume == "" {
		c.Volume = "+0%"
	}
}

func makeWsHeaders() http.Header {
	header := make(http.Header)
	header.Set("Pragma", "no-cache")
	header.Set("Cache-Control", "no-cache")
	header.Set("Origin", "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold")
	header.Set("Accept-Encoding", "gzip, deflate, br")
	header.Set("Accept-Language", "en-US,en;q=0.9")
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.77 Safari/537.36 Edg/91.0.864.41")
	return header
}

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
	data = escapeReplacer.Replace(data)
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

type audioData struct {
	Data  []byte
	Index int
}

// WriteStreamTo  write audio stream to io.WriteCloser
func (c *Communicate) WriteStreamTo(rc io.Writer) error {
	op, err := c.stream()
	if err != nil {
		return err
	}
	audioBinaryData := make([][][]byte, c.audioDataIndex)
	for data := range op {
		if _, ok := data["end"]; ok {
			if len(audioBinaryData) == c.audioDataIndex {
				break
			}
		}
		if t, ok := data["type"]; ok && t == "audio" {
			data := data["data"].(audioData)
			audioBinaryData[data.Index] = append(audioBinaryData[data.Index], data.Data)
		}
		if e, ok := data["error"]; ok {
			fmt.Printf("has error err: %v\n", e)
		}
	}

	for _, dataSlice := range audioBinaryData {
		for _, data := range dataSlice {
			rc.Write(data)
		}
	}
	return nil
}

func (c *Communicate) stream() (<-chan map[string]interface{}, error) {
	texts := splitTextByByteLength(
		escape(removeIncompatibleCharacters(c.text)),
		calculateMaxMessageSize(c.opt.Pitch, c.opt.Voice, c.opt.Rate, c.opt.Volume),
	)
	c.audioDataIndex = len(texts)

	c.finalUtterance = make(map[int]int)
	c.prevIdx = -1
	c.shiftTime = -1

	output := make(chan map[string]interface{})

	for idx, text := range texts {
		wsURL := constant.EdgeWssEndpoint + "&ConnectionId=" + generateConnectID()
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(wsURL, makeWsHeaders())
		if err != nil {
			return nil, err
		}

		currentTime := currentTimeInMST()
		err = c.sendConfig(conn, currentTime)
		if err != nil {
			conn.Close()
			return nil, err
		}
		err = c.sendSSML(conn, currentTime, text)
		if err != nil {
			conn.Close()
			return nil, err
		}

		go c.handleStream(conn, output, idx)
	}
	c.op = output
	return output, nil
}

// Sends the request to the service.
func (c *Communicate) sendConfig(conn *websocket.Conn, currentTime string) error {
	// Prepare the request to be sent to the service.
	//
	// Note sentenceBoundaryEnabled and wordBoundaryEnabled are actually supposed
	// to be booleans, but Edge Browser seems to send them as strings.
	//
	// This is a bug in Edge as Azure Cognitive Services actually sends them as
	// bool and not string. For now I will send them as bool unless it causes
	// any problems.
	//
	// Also pay close attention to double { } in request (escape for f-string).
	return conn.WriteMessage(websocket.TextMessage, []byte(
		"X-Timestamp:"+currentTime+"\r\n"+
			"Content-Type:application/json; charset=utf-8\r\n"+
			"Path:speech.config\r\n\r\n"+
			`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":false,"wordBoundaryEnabled":true},"outputFormat":"audio-24khz-48kbitrate-mono-mp3"}}}}`+"\r\n",
	))
}

func (c *Communicate) sendSSML(conn *websocket.Conn, currentTime string, text []byte) error {
	return conn.WriteMessage(websocket.TextMessage,
		[]byte(
			ssmlHeadersAppendExtraData(
				generateConnectID(),
				currentTime,
				makeSsml(string(text), c.opt.Pitch, c.opt.Voice, c.opt.Rate, c.opt.Volume),
			),
		))
}

type unknownResponse struct {
	Message string
}

type unexpectedResponse struct {
	Message string
}

type noAudioReceived struct {
	Message string
}

type webSocketError struct {
	Message string
}

func (c *Communicate) handleStream(conn *websocket.Conn, output chan map[string]interface{}, idx int) {
	// download indicates whether we should be expecting audio data,
	// this is so what we avoid getting binary data from the websocket
	// and falsely thinking it's audio data.
	downloadAudio := false

	// audioWasReceived indicates whether we have received audio data
	// from the websocket. This is so we can raise an exception if we
	// don't receive any audio data.
	audioWasReceived := false

	defer conn.Close()
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("communicate.Stream recovered from panic: %v stack: %s", err, string(debug.Stack()))
		}
	}()

	for {
		msgType, message, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				// WebSocket error
				output <- map[string]interface{}{
					"error": webSocketError{Message: err.Error()},
				}
			}
			break
		}
		switch msgType {
		case websocket.TextMessage:
			parameters, data := processWebsocketTextMessage(message)
			path := parameters["Path"]

			switch path {
			case "turn.start":
				downloadAudio = true
			case "turn.end":
				output <- map[string]interface{}{
					"end": "",
				}
				downloadAudio = false
				break // End of audio data

			case "audio.metadata":
				meta, err := getMetaHubFrom(data)
				if err != nil {
					output <- map[string]interface{}{
						"error": unknownResponse{Message: err.Error()},
					}
					break
				}

				for _, metaObj := range meta.Metadata {
					metaType := metaObj.Type
					if idx != c.prevIdx {
						c.shiftTime = sumWithMap(idx, c.finalUtterance)
						c.prevIdx = idx
					}
					switch metaType {
					case "WordBoundary":
						c.finalUtterance[idx] = metaObj.Data.Offset + metaObj.Data.Duration + 8_750_000
						output <- map[string]interface{}{
							"type":     metaType,
							"offset":   metaObj.Data.Offset + c.shiftTime,
							"duration": metaObj.Data.Duration,
							"text":     metaObj.Data.Text,
						}
					case "SessionEnd":
						// do nothing
					default:
						output <- map[string]interface{}{
							"error": unknownResponse{Message: "Unknown metadata type: " + metaType},
						}
						break
					}
				}
			case "response":
				// do nothing
			default:
				output <- map[string]interface{}{
					"error": unknownResponse{Message: "The response from the service is not recognized.\n" + string(message)},
				}
				break
			}
		case websocket.BinaryMessage:
			if !downloadAudio {
				output <- map[string]interface{}{
					"error": unknownResponse{"We received a binary message, but we are not expecting one."},
				}
			}

			if len(message) < 2 {
				output <- map[string]interface{}{
					"error": unknownResponse{"We received a binary message, but it is missing the header length."},
				}
			}

			headerLength := int(binary.BigEndian.Uint16(message[:2]))
			if len(message) < headerLength+2 {
				output <- map[string]interface{}{
					"error": unknownResponse{"We received a binary message, but it is missing the audio data."},
				}
			}

			audioBinaryData := message[headerLength+2:]
			output <- map[string]interface{}{
				"type": "audio",
				"data": audioData{
					Data:  audioBinaryData,
					Index: idx,
				},
			}
			audioWasReceived = true
		default:
			if message != nil {
				output <- map[string]interface{}{
					"error": webSocketError{
						Message: string(message),
					},
				}
			} else {
				output <- map[string]interface{}{
					"error": webSocketError{
						Message: "Unknown error",
					},
				}
			}
		}

	}
	if !audioWasReceived {
		output <- map[string]interface{}{
			"error": noAudioReceived{Message: "No audio was received. Please verify that your parameters are correct."},
		}
	}
}

func sumWithMap(idx int, m map[int]int) int {
	sumResult := 0
	for i := 0; i < idx; i++ {
		sumResult += m[i]
	}
	return sumResult
}

type textEntry struct {
	Text         string `json:"text"`
	BoundaryType string `json:"BoundaryType"`
	Length       int64  `json:"Length"`
}
type dataEntry struct {
	Offset   int       `json:"Offset"`
	Duration int       `json:"Duration"`
	Text     textEntry `json:"text"`
}
type metaDataEntry struct {
	Type string    `json:"Type"`
	Data dataEntry `json:"Data"`
}
type metaHub struct {
	Metadata []metaDataEntry `json:"Metadata"`
}

func getMetaHubFrom(data []byte) (*metaHub, error) {
	metadata := &metaHub{}
	err := json.Unmarshal(data, &metadata)
	if err != nil {
		return nil, fmt.Errorf("err=%s, data=%s", err.Error(), string(data))
	}
	return metadata, nil
}

func processWebsocketTextMessage(data []byte) (map[string]string, []byte) {
	headers := make(map[string]string)
	headerEndIndex := bytes.Index(data, []byte("\r\n\r\n"))
	headerLines := bytes.Split(data[:headerEndIndex], []byte("\r\n"))

	for _, line := range headerLines {
		header := bytes.SplitN(line, []byte(":"), 2)
		if len(header) == 2 {
			headers[string(bytes.TrimSpace(header[0]))] = string(bytes.TrimSpace(header[1]))
		}
	}

	return headers, data[headerEndIndex+4:]
}

func removeIncompatibleCharacters(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return ' '
		}
		return r
	}, str)
}

func generateConnectID() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func calculateMaxMessageSize(pitch, voice string, rate string, volume string) int {
	websocketMaxSize := 1 << 16
	overheadPerMessage := len(ssmlHeadersAppendExtraData(generateConnectID(), currentTimeInMST(), makeSsml("", pitch, voice, rate, volume))) + 50
	return websocketMaxSize - overheadPerMessage
}

func currentTimeInMST() string {
	// Use time.FixedZone to represent a fixed timezone offset of 0 (UTC)
	zone := time.FixedZone("UTC", 0)
	now := time.Now().In(zone)
	return now.Format("Mon Jan 02 2006 15:04:05 GMT-0700 (MST)")
}

type Speak struct {
	XMLName xml.Name `xml:"speak"`
	Version string   `xml:"version,attr"`
	Xmlns   string   `xml:"xmlns,attr"`
	Lang    string   `xml:"xml:lang,attr"`
	Voice   []Voice  `xml:"voice"`
}

type Voice struct {
	Name    string  `xml:"name,attr"`
	Prosody Prosody `xml:"prosody"`
}
type Prosody struct {
	// Contour represents changes in pitch. These changes are represented as an array of targets at specified time
	//positions in the speech output. Sets of parameter pairs define each target. For example:
	//
	//<prosody contour="(0%,+20Hz) (10%,-2st) (40%,+10Hz)">
	//
	//The first value in each set of parameters specifies the location of the pitch change as a percentage of the
	//duration of the text. The second value specifies the amount to raise or lower the pitch by using a relative
	//value or an enumeration value for pitch (see pitch).
	Contour string `xml:"contour,attr,omitempty"`
	//Indicates the baseline pitch for the text. Pitch changes can be applied at the sentence level. The pitch changes
	//should be within 0.5 to 1.5 times the original audio. You can express the pitch as:
	//An absolute value:
	//Expressed as a number followed by "Hz" (Hertz). For example, <prosody pitch="600Hz">some text</prosody>.
	//A relative value:
	//	As a relative number: Expressed as a number preceded by "+" or "-" and followed by "Hz" or "st" that specifies
	//	an amount to change the pitch. For example:
	//	<prosody pitch="+80Hz">some text</prosody> or <prosody pitch="-2st">some text</prosody>.
	//	The "st" indicates the change unit is semitone, which is half of a tone (a half step) on the standard diatonic scale.
	//As a percentage: Expressed as a number preceded by "+" (optionally) or "-" and followed by "%", indicating the
	//relative change. For example: <prosody pitch="50%">some text</prosody> or <prosody pitch="-50%">some text</prosody>.
	// A constant value:
	//	x-low
	//	low
	//	medium
	//	high
	//	x-high
	//	default
	Pitch string `xml:"pitch,attr"`
	// Indicates the speaking rate of the text. Speaking rate can be applied at the word or sentence level. The rate changes
	//should be within 0.5 to 2 times the original audio. You can express rate as:
	//A relative value:
	//	As a relative number: Expressed as a number that acts as a multiplier of the default. For example, a value of 1 results
	//	in no change in the original rate. A value of 0.5 results in a halving of the original rate. A value of 2 results in
	//	twice the original rate.
	//	As a percentage: Expressed as a number preceded by "+" (optionally) or "-" and followed by "%", indicating the relative
	//	change. For example:
	//	<prosody rate="50%">some text</prosody> or <prosody rate="-50%">some text</prosody>.
	//	A constant value:
	//	x-slow
	//	slow
	//	medium
	//	fast
	//	x-fast
	//	default
	Rate string `xml:"rate,attr"`
	// Indicates the volume level of the speaking voice. Volume changes can be applied at the sentence level. You can express
	//the volume as:
	// An absolute value: Expressed as a number in the range of 0.0 to 100.0, from quietest to loudest, such as 75.
	//The default value is 100.0.
	// A relative value:
	// As a relative number: Expressed as a number preceded by "+" or "-" that specifies an amount to change the volume.
	//Examples are +10 or -5.5.
	// As a percentage: Expressed as a number preceded by "+" (optionally) or "-" and followed by "%", indicating the
	//relative change. For example:
	//	<prosody volume="50%">some text</prosody> or <prosody volume="+3%">some text</prosody>.
	//
	//	A constant value:
	//	silent
	//	x-soft
	//	soft
	//	medium
	//	loud
	//	x-loud
	//	default
	Volume string `xml:"volume,attr"`
	Text   string `xml:",chardata"`
}

func makeSsml(text string, pitch, voice string, rate string, volume string) string {
	ssml := &Speak{
		XMLName: xml.Name{Local: "speak"},
		Version: "1.0",
		Xmlns:   "http://www.w3.org/2001/10/synthesis",
		Lang:    "en-US",
		Voice: []Voice{{
			Name: voice,
			Prosody: Prosody{
				Pitch:  pitch,
				Rate:   rate,
				Volume: volume,
				Text:   text,
			},
		}},
	}

	output, err := xml.MarshalIndent(ssml, "", "  ")
	if err != nil {
		return ""
	}
	return string(output)
}

func ssmlHeadersAppendExtraData(requestID string, timestamp string, ssml string) string {
	ssmlHeaderTemplate := "X-RequestId:%s\r\nContent-Type:application/ssml+xml\r\nX-Timestamp:%sZ\r\nPath:ssml\r\n\r\n"
	headers := fmt.Sprintf(
		ssmlHeaderTemplate,
		requestID,
		timestamp,
	)
	return headers + ssml
}
