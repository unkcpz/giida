package giida

import (
  "fmt"
	"time"
	"reflect"
)

// Tasker is a unit run a Execute
type Tasker interface {
  Execute()
}

type Processer interface{
  Run() Wait
}

type Process struct {
  Name string
  task  Tasker
	InPorts map[string]reflect.Value
	OutPorts map[string]reflect.Value
	inPool map[string]reflect.Value
	inChannel map[string]reflect.Value
}

func NewProcess(name string, task Tasker) *Process {
  p := &Process{
    Name: name,
    task: task,
		InPorts: make(map[string]reflect.Value),
		OutPorts: make(map[string]reflect.Value),
		inPool: make(map[string]reflect.Value),
		inChannel: make(map[string]reflect.Value),
  }
	mapPort(p)
  return p
}

func mapPort(p *Process) {
	// Set value to task's fields
	val := reflect.ValueOf(p.task).Elem()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := field.Type()
		fieldName := val.Type().Field(i).Name
		if fieldType.ChanDir() == reflect.RecvDir {
			p.InPorts[fieldName] = val.FieldByName(fieldName)
		}
		if fieldType.ChanDir() == reflect.SendDir {
			p.OutPorts[fieldName] = val.FieldByName(fieldName)
		}
	}
}

func (p *Process) In(portName string, v interface{}) {
	p.inPool[portName] = reflect.ValueOf(v)
	// fmt.Println(reflect.TypeOf(v).ChanDir())
  chanType := reflect.ChanOf(reflect.BothDir, reflect.TypeOf(v))
  p.inChannel[portName] = reflect.MakeChan(chanType, 0)

	p.InPorts[portName].Set(p.inChannel[portName])
}

func (p *Process) Out(portName string, v interface{}) {
	p.OutPorts[portName].Set(reflect.ValueOf(v))
}

// Done notifies that the process is finished
type Done struct{}

// Wait is a channel signalling of a completion
type Wait chan struct{}

func (p *Process) Run() Wait {
	t := p.task
	wait := make(Wait)
	go func() {
		// fmt.Printf("%s | Running %s\n", timeStamp(), p.Name)
		t.Execute()

		wait <- Done{}
		// fmt.Printf("%s | %s Finished\n", timeStamp(), p.Name)
	}()
	for k, v := range p.inPool {
		p.inChannel[k].Send(v)
	}
	return wait
}

func timeStamp() string {
	t := time.Now()
	return fmt.Sprintf(t.Format("2006/01-02/15:04:05"))
}

// type InputGuard struct {
//   ports map[string]bool
//   complete int
// }
//
// //
// func NewInputGuard(ports ...string) *InputGuard {
//   portMap := make(map[string]bool, len(ports))
//   for _, p := range ports {
//     portMap[p] = false
//   }
//   return &InputGuard{portMap, 0}
// }
//
// // Complete is called when a port is closed and returns true when all the ports have been closed
// func (g *InputGuard) Complete(port string) bool {
//   if !g.ports[port] {
//     g.ports[port] = true
//     g.complete++
//   }
//   return g.complete >= len(g.ports)
// }
