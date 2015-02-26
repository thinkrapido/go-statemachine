package statemachine_test

import (
	. "github.com/thinkrapido/go-statemachine/state"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

  "time"
)

func testState(toBe string, done Done, after func() (*Machine)) {
    c := make(chan string, 0)

    sm1 := after()

    go func(c chan string) {
      time.Sleep(time.Second / 10)
      c <- sm1.CurrentState()
    }(c)

    Expect(<-c).To(Equal(toBe))

    close(done)
}

var _ = Describe("Statemachine", func() {

  var (
    sm1 *Machine
  )

  BeforeEach(func() {

    sm1 = NewMachine()

    sm1.Learn("state 1", "state 2", "walk")
    sm1.Learn("state 2", "state 3", "walk")
    sm1.Learn("state 3", "state 1", "walk")

    sm1.Learn("state 2", "state 2", "stay")

    sm1.SetStartState("state 1")
    sm1.Run()
  })

  It("should be state 1", func() {
    Expect(sm1.CurrentState()).To(Equal("state 1"))
  })

  Describe("traversion", func() {
    Context("After Trigger 'walk'", func() {
      It("should be 'state 2'", func(done Done) {
        testState("state 2", done, func() (*Machine) {
          return sm1.Trigger("walk")
        })
      }, .2)

      It("should be 'state 3'", func(done Done) {
        testState("state 3", done, func() (*Machine) {
          return sm1.Trigger("walk").Trigger("walk")
        })
      }, .2)

      It("should be 'state 1' again", func(done Done) {
        testState("state 1", done, func() (*Machine) {
          return sm1.Trigger("walk").Trigger("walk").Trigger("walk")
        })
      })

      It("should stay on 'state 2'", func(done Done) {
        testState("state 2", done, func() (*Machine) {
          return sm1.Trigger("walk").Trigger("stay")
        })
      })

    })
  })

})

type notifiable struct {
  inconsistency *Event
}
func (n *notifiable) Notify(e *Event) {
  switch e.Event {
    case InconsistencyEvent:
      n.inconsistency = e
  }
}

var _ = Describe("Statemachine", func() {

  var (
    sm1 *Machine
    actionExecuted bool
    noti *notifiable
  )

  BeforeEach(func() {

    noti = &notifiable{}

    sm1 = NewMachine()

    sm1.Learn("state 1", "state 2", "walk", func() {
      actionExecuted = true
    })

    sm1.AddListener(noti)
    
    sm1.SetStartState("state 1")
    sm1.Run()
  })

  Describe("action method", func() {
    It("should be triggered", func(done Done) {
      c := make(chan string, 0)

      go func(c chan string) {
        sm1.Trigger("walk")
        time.Sleep(time.Second / 10)
        c <- sm1.CurrentState()
      }(c)

      Expect(<-c).To(Equal("state 2"))
      Expect(actionExecuted).To(Equal(true))
      Expect(noti.inconsistency).To(BeNil())

      close(done)
    })
  })

})

var _ = Describe("Statemachine", func() {

  var (
    sm1 *Machine
    actionExecuted bool
    recoverExecuted bool
    noti *notifiable
  )

  BeforeEach(func() {

    noti = &notifiable{}

    sm1 = NewMachine()

    sm1.Learn("state 1", "state 2", "walk", func() {
      panic("Action")
      actionExecuted = true
    }, func() {
      recoverExecuted = true
    })

    sm1.AddListener(noti)
    
    sm1.SetStartState("state 1")
    sm1.Run()
  })

  Describe("recovery method", func() {
    It("should be triggered", func(done Done) {
      c := make(chan string, 0)

      go func(c chan string) {
        sm1.Trigger("walk")
        time.Sleep(time.Second / 10)
        c <- sm1.CurrentState()
      }(c)

      Expect(<-c).To(Equal("state 2"))
      Expect(actionExecuted).To(Equal(false))
      Expect(recoverExecuted).To(Equal(true))
      Expect(noti.inconsistency).To(BeNil())

      close(done)
    })
  })

})

var _ = Describe("Statemachine", func() {

  var (
    sm1 *Machine
    actionExecuted bool
    recoverExecuted bool
    noti *notifiable
  )

  BeforeEach(func() {

    noti = &notifiable{}

    sm1 = NewMachine()

    sm1.Learn("state 1", "state 2", "walk", func() {
      panic("Action")
      actionExecuted = true
    }, func() {
      panic("Recovery")
      recoverExecuted = true
    })

    sm1.AddListener(noti)
    
    sm1.SetStartState("state 1")
    sm1.Run()
  })

  Describe("not recoverd", func() {
    It("should be triggered", func(done Done) {
      c := make(chan string, 0)

      go func(c chan string) {
        sm1.Trigger("walk")
        time.Sleep(time.Second / 10)
        c <- sm1.CurrentState()
      }(c)

      Expect(<-c).To(Equal("state 1"))
      Expect(actionExecuted).To(Equal(false))
      Expect(recoverExecuted).To(Equal(false))
      Expect(noti.inconsistency).ShouldNot(BeNil())
      Expect(noti.inconsistency.Trigger).To(Equal("walk"))
      Expect(noti.inconsistency.Message).To(Equal("Recover function failed.\n\t\tRecovery"))

      close(done)
    })
  })
})

var _ = Describe("Statemachine", func() {

  var (
    sm1 *Machine
    actionExecuted bool
    recoverExecuted bool
    noti *notifiable
  )

  BeforeEach(func() {

    noti = &notifiable{}

    sm1 = NewMachine()

    sm1.Learn("state 1", "state 2", "walk", func() {
      panic("Action")
      actionExecuted = true
    })

    sm1.AddListener(noti)
    
    sm1.SetStartState("state 1")
    sm1.Run()
  })

  Describe("recovery method not provided", func() {
    It("should be triggered", func(done Done) {
      c := make(chan string, 0)

      go func(c chan string) {
        sm1.Trigger("walk")
        time.Sleep(time.Second / 10)
        c <- sm1.CurrentState()
      }(c)

      Expect(<-c).To(Equal("state 1"))
      Expect(actionExecuted).To(Equal(false))
      Expect(recoverExecuted).To(Equal(false))
      Expect(noti.inconsistency).ShouldNot(BeNil())
      Expect(noti.inconsistency.Trigger).To(Equal("walk"))
      Expect(noti.inconsistency.Message).To(Equal("No recover function provided.\n\t\tAction"))

      close(done)
    })
  })
})

