package process

import "WarpGPT/pkg/requestbody"

type Process struct {
	Conversation requestbody.Conversation
}

type ConversationProcesser interface {
	ProcessMethod()
}

func Do(p ConversationProcesser) {
	p.ProcessMethod()
}
