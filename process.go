package flowmat

import (
  "reflect"
  "sync"
  // "fmt"
)

type Tasker interface{
  Execute()
}

type Process struct {
  name string
  task Tasker
  InPorts map[string]*Port
  OutPorts map[string]*Port
  ports map[string]*Port
  exposePorts map[string]*Port
}

// NewProcess create a Process of a task
func NewProcess(name string, task Tasker) *Process {
  proc := &Process{
    name: name,
    task: task,
    InPorts: make(map[string]*Port),
    OutPorts: make(map[string]*Port),
    ports: make(map[string]*Port),
    exposePorts: make(map[string]*Port),
  }
  val := reflect.ValueOf(task).Elem()
  // Bind every field of task to a port
  for i:=0; i<val.NumField(); i++ {
    fieldName := val.Type().Field(i).Name
    proc.ports[fieldName] = &Port{
      channel: make(chan interface{}),
    }
  }

  return proc
}

// Name return process's name
func (p  *Process) Name() string {
  return p.name
}

// ExposeIn return Port for channel
func (p *Process) ExposeIn(name string) *Port {
  port := p.ports[name]
  p.InPorts[name] = port
  p.exposePorts[name] = port
  return port
}

// ExposeOut return Port for channel
func (p *Process) ExposeOut(name string) *Port {
  port := p.ports[name]
  p.OutPorts[name] = port
  p.exposePorts[name] = port
  return port
}


// In set InPort's cache
func (p *Process) In(name string, data interface{}) {
  port := p.ports[name]
  port.cache = data
  p.InPorts[name] = port
}

// Out get OutPort's cache
func (p *Process) Out(name string) interface{} {
  port := p.ports[name]
  p.OutPorts[name] = port

  return port.cache
}

// Run process
func (p *Process) Run() {
  task := p.task
  val := reflect.ValueOf(task).Elem()

  unset := make([]*Port, 0)
  for name, _ := range p.ports {
    _, okIn := p.InPorts[name]
    _, okOut := p.OutPorts[name]
    if !okIn && !okOut {
      unset = append(unset, p.ExposeOut(name))
    }
  }

  // gorountine get input and run Execute()
  go func(){
    var wg sync.WaitGroup
    for name, port := range p.InPorts {
      wg.Add(1)
      go func(name string, port *Port) {
        defer wg.Done()
        ch := port.channel
        v := reflect.ValueOf(<-ch)
        val.FieldByName(name).Set(v)
        if port.cache == nil{
          port.cache = v.Interface()
        }
        close(ch)
      }(name, port)
    }
    wg.Wait()

    // Execute the function of Process
    task.Execute()

    for name, port := range p.OutPorts {
      wg.Add(1)
      go func(name string, port *Port) {
        defer wg.Done()
        ch := port.channel
        v := val.FieldByName(name).Interface()
        port.cache = v
        ch <- v
        close(ch)
      }(name, port)
    }
    wg.Wait()
  }()

  // Feed the inputs aka start Chain
  for name, port := range p.InPorts {
    if _, ok := p.exposePorts[name]; !ok {
      port.Feed(nil)
    }
  }
  // Extract the outputs
  for _, port := range unset {
    port.Extract()
  }
}
