package progress

// ChannelReporter 将进度写入 channel（供 TUI bubbletea 消费）
type ChannelReporter struct {
	Ch chan<- Event
}

// NewChannelReporter 创建 channel reporter
func NewChannelReporter(ch chan<- Event) *ChannelReporter {
	return &ChannelReporter{Ch: ch}
}

func (c *ChannelReporter) Report(e Event) {
	if c == nil || c.Ch == nil {
		return
	}
	select {
	case c.Ch <- e:
	default:
	}
}
