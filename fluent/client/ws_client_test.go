package client_test

import (
	"bytes"
	"errors"

	"time"

	. "github.com/IBM/fluent-forward-go/fluent/client"
	"github.com/IBM/fluent-forward-go/fluent/client/clientfakes"
	"github.com/IBM/fluent-forward-go/fluent/client/ws/ext"
	"github.com/IBM/fluent-forward-go/fluent/client/ws/ext/extfakes"
	"github.com/IBM/fluent-forward-go/fluent/client/ws/wsfakes"
	"github.com/IBM/fluent-forward-go/fluent/protocol"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IAMAuthInfo", func() {
	It("gets and sets an IAM token", func() {
		iai := NewIAMAuthInfo("a")
		Expect(iai.IAMToken()).To(Equal("a"))
		iai.SetIAMToken("b")
		Expect(iai.IAMToken()).To(Equal("b"))
	})
})

var _ = Describe("WSClient", func() {
	var (
		factory    *clientfakes.FakeWSConnectionFactory
		client     *WSClient
		clientSide ext.Conn
		conn       *wsfakes.FakeConnection
		session    *WSSession
	)

	BeforeEach(func() {
		factory = &clientfakes.FakeWSConnectionFactory{}
		client = &WSClient{
			ConnectionFactory: factory,
		}
		clientSide = &extfakes.FakeConn{}
		conn = &wsfakes.FakeConnection{}
		session = &WSSession{Connection: conn}

		Expect(factory.NewCallCount()).To(Equal(0))
		Expect(client.Session).To(BeNil())
	})

	JustBeforeEach(func() {
		factory.NewReturns(clientSide, nil)
		factory.NewSessionReturns(session)
	})

	Describe("Connect", func() {
		It("Does not return an error", func() {
			Expect(client.Connect()).ToNot(HaveOccurred())
		})

		It("Gets the connection from the ConnectionFactory", func() {
			err := client.Connect()
			Expect(err).NotTo(HaveOccurred())
			Expect(factory.NewCallCount()).To(Equal(1))
			Expect(factory.NewSessionCallCount()).To(Equal(1))
			Expect(client.Session).To(Equal(session))
			Expect(client.Session.Connection).To(Equal(conn))
		})

		When("the factory returns an error", func() {
			var (
				connectionError error
			)

			JustBeforeEach(func() {
				connectionError = errors.New("Nope")
				factory.NewReturns(nil, connectionError)
			})

			It("Returns an error", func() {
				err := client.Connect()
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeIdenticalTo(connectionError))
			})
		})

		When("the factory returns an error", func() {
			var (
				connectionError error
			)

			JustBeforeEach(func() {
				connectionError = errors.New("Nope")
				factory.NewReturns(nil, connectionError)
			})

			It("Returns an error", func() {
				err := client.Connect()
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeIdenticalTo(connectionError))
			})
		})
	})

	Describe("Disconnect", func() {
		When("the session is not nil", func() {
			JustBeforeEach(func() {
				err := client.Connect()
				Expect(err).NotTo(HaveOccurred())
				time.Sleep(100 * time.Millisecond)
			})

			It("closes the connection", func() {
				Expect(client.Disconnect()).ToNot(HaveOccurred())
				Expect(conn.CloseCallCount()).To(Equal(1))
			})
		})

		When("the session is nil", func() {
			JustBeforeEach(func() {
				client.Session = nil
			})

			It("does not error or panic", func() {
				Expect(func() {
					Expect(client.Disconnect()).ToNot(HaveOccurred())
				}).ToNot(Panic())
			})
		})
	})

	Describe("Reconnect", func() {
		JustBeforeEach(func() {
			err := client.Connect()
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
		})

		It("calls Disconnect and creates a new Session", func() {
			Expect(client.Reconnect()).ToNot(HaveOccurred())

			Expect(conn.CloseCallCount()).To(Equal(1))

			Expect(factory.NewSessionCallCount()).To(Equal(2))
			Expect(client.Session.Connection).ToNot(BeNil())
		})
	})

	Describe("SendMessage", func() {
		var (
			msg protocol.MessageExt
		)

		BeforeEach(func() {
			msg = protocol.MessageExt{
				Tag:       "foo.bar",
				Timestamp: protocol.EventTime{time.Now()}, //nolint
				Record:    map[string]interface{}{},
				Options:   &protocol.MessageOptions{},
			}
		})

		JustBeforeEach(func() {
			err := client.Connect()
			Expect(err).ToNot(HaveOccurred())
			time.Sleep(100 * time.Millisecond)
		})

		It("Sends the message", func() {
			bits, _ := msg.MarshalMsg(nil)
			Expect(client.SendMessage(&msg)).ToNot(HaveOccurred())

			writtenbits := conn.WriteArgsForCall(0)
			Expect(bytes.Equal(bits, writtenbits)).To(BeTrue())
		})

		When("the connection is disconnected", func() {
			JustBeforeEach(func() {
				err := client.Disconnect()
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns an error", func() {
				Expect(client.SendMessage(&msg)).To(MatchError("no active session"))
			})
		})

		When("the connection is closed with an error", func() {
			BeforeEach(func() {
				conn.ListenReturns(errors.New("BOOM"))
			})

			It("returns the error", func() {
				Expect(client.SendMessage(&msg)).To(MatchError("BOOM"))
			})
		})
	})
})
