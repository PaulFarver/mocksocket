package producer

import (
	"context"
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

var (
	IN_PROGRESS_MESSAGE = Message{
		Status: STATUS_IN_PROGRESS,
		Result: nil,
		Error:  nil,
	}
	SUCCESS_MESSAGE = Message{
		Status: STATUS_DONE,
		Result: &Result{
			SignedDownloadUrl: "http://africau.edu/images/default/sample.pdf",
		},
		Error: nil,
	}
	FAILED_MESSAGE = Message{
		Status: STATUS_FAILED,
		Result: nil,
		Error: &ErrorResponse{
			Code:    500,
			Message: "Something went wrong while generating zip file",
		},
	}
	ErrNoMoreMessages = errors.New("no more messages")
)

type Producer struct {
	Delay    time.Duration
	Sequence func() (Message, error)
}

const (
	STATUS_NOT_STARTED = "NOT_STARTED"
	STATUS_IN_PROGRESS = "IN_PROGRESS"
	STATUS_DONE        = "DONE"
	STATUS_FAILED      = "FAILED"
)

type Result struct {
	SignedDownloadUrl string `json:"signedDownloadUrl"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Message struct {
	Status string         `json:"status"`
	Result *Result        `json:"result,omitempty"`
	Error  *ErrorResponse `json:"error,omitempty"`
}

func (p *Producer) Produce() (chan Message, func()) {
	ch := make(chan Message)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(ch)
				return
			case <-time.After(p.Delay):
				msg, err := p.Sequence()
				if err == nil {
					ch <- msg
				}
			}
		}
	}()
	return ch, cancel
}

func generateMessage() Message {
	switch rand.Intn(3) {
	case 0:
		return SUCCESS_MESSAGE
	case 1:
		return FAILED_MESSAGE
	default:
		return IN_PROGRESS_MESSAGE
	}
}

func GetSequence(seq string) (func() (Message, error), error) {
	switch seq {
	case "fail":
		msgs := []Message{
			IN_PROGRESS_MESSAGE,
			IN_PROGRESS_MESSAGE,
			FAILED_MESSAGE,
		}
		i := -1
		return func() (Message, error) {
			i = i + 1
			if len(msgs) <= i {
				return Message{}, errors.New("sequence completed")
			}
			return msgs[i], nil
		}, nil
	case "success":
		msgs := []Message{
			IN_PROGRESS_MESSAGE,
			IN_PROGRESS_MESSAGE,
			SUCCESS_MESSAGE,
		}
		i := -1
		return func() (Message, error) {
			i = i + 1
			if len(msgs) <= i {
				return Message{}, errors.New("sequence completed")
			}
			return msgs[i], nil
		}, nil
	case "random":
		return func() (Message, error) {
			return generateMessage(), nil
		}, nil
	default:
		return func() (Message, error) {
			return generateMessage(), nil
		}, errors.Errorf("invalid sequence '%s'", seq)

	}
}
