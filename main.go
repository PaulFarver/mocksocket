package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/paulfarver/mocksocket/producer"
	"github.com/sirupsen/logrus"
)

type WebsocketHandler struct {
	upgrader *websocket.Upgrader
	logger   *logrus.Logger
}

func main() {
	w := WebsocketHandler{
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		// producer: producer.Producer{
		// 	Delay: time.Second,
		// },
		logger: logrus.New(),
	}

	e := echo.New()
	e.GET("/sequence/:sequence/delay/:delay", w.HandleWebsocket)
	e.Logger.Fatal(e.Start(":8080"))
}

type Request struct {
	Sequence string `param:"sequence"`
	Delay    string `param:"delay"`
}

func (w *WebsocketHandler) HandleWebsocket(ctx echo.Context) error {
	req := Request{}
	if err := ctx.Bind(&req); err != nil {
		return err
	}

	delay, err := time.ParseDuration(req.Delay)
	if err != nil {
		return err
	}

	w.logger.Infof("Handling websocket request for sequence %s and delay %s", req.Sequence, delay)

	conn, err := w.upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		w.logger.WithError(err).Error("Failed to upgrade connection")

		return err
	}
	defer conn.Close()

	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Println("close handler")
		return nil
	})

	c, cancel := context.WithCancel(ctx.Request().Context())
	defer cancel()

	seq, err := producer.GetSequence(req.Sequence)
	if err != nil {
		w.logger.WithError(err).Warn("failed to get sequence. Defaulting to random")
	}

	producer := producer.Producer{
		Delay:    delay,
		Sequence: seq,
	}

	ch, closer := producer.Produce()
	defer closer()

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				return
			}
		}
	}()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				w.logger.Info("producer channel closed")

				return nil
			}
			w.logger.WithField("msg", msg).Info("Sending message to client")
			if err := conn.WriteJSON(msg); err != nil {
				w.logger.WithError(err).Error("Error sending message to client")

				return err
			}
		case <-c.Done():
			w.logger.Info("context canceled")
			return nil
		}
	}
}
