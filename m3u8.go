package m3u8

type Playlist struct {
	EndOfList      bool
	Version        int
	TargetDuration int
	ListType       string
	MediaSequence  int
	Segments       []Segment
	Variants       []Variant
	Media          []Media
}

func (p Playlist) IsMaster() (b bool) {
	if len(p.Variants) > 0 {
		b = true
	}
	return
}

type Map struct {
	Url    string
	Length int
	Offset int
}

type Key struct {
	Method string
	Url    string
	IV     string
}

type Segment struct {
	Duration float64
	Url      string
	Title    string
	Length   int
	Offset   int
	Bitrate  int64
	Key      *Key // optional
	Map      *Map // optional
}

type Media struct {
	Type       string
	Url        string
	GroupID    string
	Language   string
	Name       string
	Default    bool
	Forced     bool
	AutoSelect bool
}

type Variant struct {
	Bandwidth        int
	AverageBandwidth int
	Url              string
	Codecs           string
	Resolution       string
	Audio            string
	Video            string
	Subtitles        string
	ClosedCaptions   string
	FrameRate        float64
}
