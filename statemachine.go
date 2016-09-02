package statemachine

import "fmt"

/**
  const definitions
*/

const (
  InconsistencyEvent = iota
  StateReachedEvent
  KillEvent
)

/** 
  type definitions
*/
type ActionFunc func()

type Machine struct {
  
  event chan string
  running bool
  
  states map[string]*state
  startState string
  currentState string

  listeners map[Controller]bool
}

type state struct {
  name string
  events map[string]*transition
}

type transition struct {
  event string
  endState string
  action ActionFunc
  recover ActionFunc
}

type Event struct {
  Machine *Machine
  Event int
  Trigger string
  Message string
}

type Controller interface {
  Notify(event *Event)
}

/**
  Machine type declaration
*/

func NewMachine() *Machine {
  out := &Machine{}
  out.Init()
  return out
}

func (sm *Machine) Init() {

    sm.event = make(chan string)
    sm.states = make(map[string]*state)
    sm.listeners = make(map[Controller]bool)
  
}

func (sm *Machine) init() {
  var start *state
  for k, s := range(sm.states) {
    if k == sm.startState {
      start = s
      break
    }
  }
  if start == nil {
    panic("no start state defined")
  }
  sm.currentState = sm.startState
  sm.notify(StateReachedEvent, sm.startState)
}

func (sm *Machine) SetStartState(state string) {
  sm.startState = state
}
func (sm *Machine) StartState() string {
  return sm.startState
}
func (sm *Machine) CurrentState() string {
  return sm.currentState
}

func (sm *Machine) Run() {
  if len(sm.states) == 0 {
    return
  }
  if sm.running {
    panic("Machine already running.")
  }

  sm.init()

  sm.running = true

  go func() {
    for event := range sm.event {
      switch event {
        case "!kill":
          sm.running = false
          close(sm.event)
          sm.notify(KillEvent, "")
          return
        default:
          state := sm.states[sm.CurrentState()]

          if t, ok := state.events[event]; ok {
            t.transit(sm)
          }
      }
    }
  }()
}

func (sm *Machine) Trigger(event string) (*Machine) {
  if event[0] == '!' {
    return sm
  }
  if !sm.running {
    panic("Machine not runnig.")
  }
  sm.event <- event
  return sm
}

func (sm *Machine) Kill() {
  if !sm.running {
    panic("Machine not runnig.")
  }
  sm.event <- "!kill"
}

func (sm *Machine) Learn(startState, endState, event string, actions ...ActionFunc) {
  s, ok := sm.states[startState]
  if !ok {
    s = &state{
      name: startState,
      events: make(map[string]*transition),
    }
    sm.states[startState] = s
  }
  var t *transition
  t, ok = s.events[event]
  if ok {
    panic("transition already exists.")
  }

  t = &transition{
    event: event,
    endState: endState,
  }
  switch len(actions) {
    case 2:
      t.recover = actions[1]
      fallthrough
    case 1:
      t.action = actions[0]
  }
  s.events[event] = t
}

func (sm *Machine) AddListener(listener Controller) {
  sm.listeners[listener] = true
}
func (sm *Machine) RemoveListener(listener Controller) {
  delete(sm.listeners, listener)
}
func (sm *Machine) notify(event int, trigger string, message ...string) {
  var msg string
  for _, m := range(message) {
    msg = m
    break
  }
  for listener, _ := range sm.listeners {
    go func() {
      listener.(Controller).Notify(&Event{sm, event, trigger, msg})
    }()
  }
}

/**
  transition type declaration
*/

func (t* transition) transit(sm *Machine) {
  if t.action != nil {
    defer func() {
      if r := recover(); r != nil {
        if t.recover != nil {
          defer func() {
            if r := recover(); r != nil {
              sm.notify(InconsistencyEvent, t.event, fmt.Sprintf("Recover function failed.\n\t\t%s", r))
            }
          }()
          t.recover()
          sm.currentState = t.endState
          sm.notify(StateReachedEvent, t.event)
        } else {
          sm.notify(InconsistencyEvent, t.event, fmt.Sprintf("No recover function provided.\n\t\t%s", r))
        }
      }
    }()
    t.action()
    sm.currentState = t.endState
    sm.notify(StateReachedEvent, t.event)
  } else {
    sm.currentState = t.endState
    sm.notify(StateReachedEvent, t.event)
  }
}